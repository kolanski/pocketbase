package jsvm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/buffer"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/process"
	"github.com/dop251/goja_nodejs/require"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/cron"
	"github.com/pocketbase/pocketbase/tools/router"
	"github.com/pocketbase/pocketbase/tools/template"
)

// LambdaFunctionPluginConfig defines the configuration for the lambda function plugin
type LambdaFunctionPluginConfig struct {
	// PoolSize specifies how many goja.Runtime instances to prewarm
	// for lambda function execution
	PoolSize int

	// MaxExecutionTime specifies the maximum execution time for lambda functions
	MaxExecutionTime time.Duration

	// MaxMemory specifies the maximum memory usage for lambda functions (in bytes)
	MaxMemory int64

	// OnInit allows custom initialization of the JS runtime
	OnInit func(vm *goja.Runtime)
}

// LambdaFunctionPlugin manages lambda function execution
type LambdaFunctionPlugin struct {
	app           core.App
	config        LambdaFunctionPluginConfig
	executors     *vmsPool
	scheduler     *cron.Cron
	router        *router.Router[*core.RequestEvent]
	httpRoutes    sync.Map // map[string]*LambdaFunctionHTTPRoute
	dbTriggers    sync.Map // map[string][]*LambdaFunctionDBTrigger
	cronJobs      sync.Map // map[string]*LambdaFunctionCronJob
	templateRegistry *template.Registry
	requireRegistry  *require.Registry
}

// LambdaFunctionHTTPRoute represents an HTTP route for an lambda function
type LambdaFunctionHTTPRoute struct {
	FunctionID string
	Method     string
	Path       string
	Handler    func(*core.RequestEvent) error
}

// LambdaFunctionDBTrigger represents a database trigger for an lambda function
type LambdaFunctionDBTrigger struct {
	FunctionID string
	Collection string
	Event      string // "create", "update", "delete"
}

// LambdaFunctionCronJob represents a cron job for an lambda function
type LambdaFunctionCronJob struct {
	FunctionID string
	Schedule   string
	JobID      string
}

// LambdaFunctionExecutionContext provides context for lambda function execution
type LambdaFunctionExecutionContext struct {
	FunctionID   string
	TriggerType  string
	Request      *http.Request
	Response     http.ResponseWriter
	Record       interface{}
	OldRecord    interface{}
	Environment  map[string]string
	StartTime    time.Time
}

// LambdaFunctionExecutionResult represents the result of lambda function execution
type LambdaFunctionExecutionResult struct {
	Success   bool
	Output    interface{}
	Error     string
	Duration  time.Duration
	Memory    int64
}

// RegisterLambdaFunctionPlugin registers the lambda function plugin with the app
func RegisterLambdaFunctionPlugin(app core.App, config LambdaFunctionPluginConfig) (*LambdaFunctionPlugin, error) {
	if config.MaxExecutionTime == 0 {
		config.MaxExecutionTime = 30 * time.Second
	}
	if config.MaxMemory == 0 {
		config.MaxMemory = 128 * 1024 * 1024 // 128MB
	}

	plugin := &LambdaFunctionPlugin{
		app:              app,
		config:           config,
		scheduler:        cron.New(),
		templateRegistry: template.NewRegistry(),
		requireRegistry:  new(require.Registry),
	}

	// Initialize VM pool
	plugin.executors = newPool(config.PoolSize, plugin.createVM)

	// Register app lifecycle hooks
	plugin.registerLifecycleHooks()

	// Load existing lambda functions after database is ready
	plugin.app.OnBootstrap().BindFunc(func(e *core.BootstrapEvent) error {
		if err := e.Next(); err != nil {
			return err
		}
		if err := plugin.loadLambdaFunctions(); err != nil {
			return err
		}
		// Register HTTP routes after functions are loaded
		plugin.registerHTTPRoutes()
		return nil
	})

	return plugin, nil
}

// createVM creates a new goja.Runtime instance for lambda function execution
func (p *LambdaFunctionPlugin) createVM() *goja.Runtime {
	vm := goja.New()

	// Enable Node.js compatibility
	p.requireRegistry.Enable(vm)
	console.Enable(vm)
	process.Enable(vm)
	buffer.Enable(vm)

	// Add PocketBase bindings
	baseBinds(vm)
	dbxBinds(vm)
	filesystemBinds(vm)
	securityBinds(vm)
	osBinds(vm)
	filepathBinds(vm)
	httpClientBinds(vm)
	formsBinds(vm)
	apisBinds(vm)
	mailsBinds(vm)

	// Add lambda function specific bindings
	vm.Set("$app", p.app)
	vm.Set("$template", p.templateRegistry)

	// Custom initialization
	if p.config.OnInit != nil {
		p.config.OnInit(vm)
	}

	return vm
}

// registerLifecycleHooks registers the necessary app lifecycle hooks
func (p *LambdaFunctionPlugin) registerLifecycleHooks() {
	// Store the router for later use and register routes
	p.app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		p.router = e.Router
		// Register HTTP routes immediately when router is available
		p.registerHTTPRoutes()
		return e.Next()
	})

	// Register database triggers
	p.registerDatabaseTriggers()

	// Start cron scheduler
	p.app.OnBootstrap().BindFunc(func(e *core.BootstrapEvent) error {
		p.scheduler.Start()
		return e.Next()
	})

	// Stop cron scheduler on termination
	p.app.OnTerminate().BindFunc(func(e *core.TerminateEvent) error {
		p.scheduler.Stop()
		return e.Next()
	})

	// Handle lambda function CRUD operations
	p.app.OnRecordCreate("lambdas").BindFunc(func(e *core.RecordEvent) error {
		if err := e.Next(); err != nil {
			return err
		}
		return p.handleFunctionCreated(e.Record)
	})

	p.app.OnRecordUpdate("lambdas").BindFunc(func(e *core.RecordEvent) error {
		if err := e.Next(); err != nil {
			return err
		}
		return p.handleFunctionUpdated(e.Record)
	})

	p.app.OnRecordDelete("lambdas").BindFunc(func(e *core.RecordEvent) error {
		if err := e.Next(); err != nil {
			return err
		}
		return p.handleFunctionDeleted(e.Record)
	})
}

// loadLambdaFunctions loads all existing lambda functions from the database
func (p *LambdaFunctionPlugin) loadLambdaFunctions() error {
	// Check if the collection exists first
	_, err := p.app.FindCollectionByNameOrId("lambdas")
	if err != nil {
		// Collection doesn't exist yet, this is fine
		p.app.Logger().Debug("Lambda functions collection not found, skipping loading")
		return nil
	}

	functions, err := p.app.FindRecordsByFilter("lambdas", "enabled = true", "", 0, 0)
	if err != nil {
		return fmt.Errorf("failed to load lambda functions: %w", err)
	}

	p.app.Logger().Info("Loading lambda functions", "count", len(functions))

	for _, function := range functions {
		p.app.Logger().Info("Registering lambda function", "name", function.GetString("name"), "id", function.Id)
		if err := p.registerFunction(function); err != nil {
			// Log error but continue loading other functions
			p.app.Logger().Error("Failed to register lambda function", "function", function.GetString("name"), "error", err)
		}
	}

	return nil
}

// registerFunction registers triggers for a specific lambda function
func (p *LambdaFunctionPlugin) registerFunction(function *core.Record) error {
	if !function.GetBool("enabled") {
		p.app.Logger().Debug("Skipping disabled function", "name", function.GetString("name"))
		return nil
	}

	functionID := function.Id
	triggers := function.GetString("triggers")
	
	p.app.Logger().Info("Processing triggers for function", "name", function.GetString("name"), "triggers", triggers)

	var triggerConfig map[string]interface{}
	if err := json.Unmarshal([]byte(triggers), &triggerConfig); err != nil {
		p.app.Logger().Error("Invalid trigger configuration", "function", function.GetString("name"), "error", err, "triggers", triggers)
		return fmt.Errorf("invalid trigger configuration: %w", err)
	}

	// Register HTTP triggers
	if httpTriggers, ok := triggerConfig["http"].([]interface{}); ok {
		p.app.Logger().Info("Found HTTP triggers", "count", len(httpTriggers), "function", function.GetString("name"))
		for _, trigger := range httpTriggers {
			if httpTrigger, ok := trigger.(map[string]interface{}); ok {
				method := strings.ToUpper(httpTrigger["method"].(string))
				path := httpTrigger["path"].(string)
				p.app.Logger().Info("Registering HTTP trigger", "method", method, "path", path, "function", function.GetString("name"))
				p.registerHTTPTrigger(functionID, method, path)
			}
		}
	} else {
		p.app.Logger().Debug("No HTTP triggers found", "function", function.GetString("name"))
	}

	// Register database triggers
	if dbTriggers, ok := triggerConfig["database"].([]interface{}); ok {
		for _, trigger := range dbTriggers {
			if dbTrigger, ok := trigger.(map[string]interface{}); ok {
				collection := dbTrigger["collection"].(string)
				event := dbTrigger["event"].(string)
				p.registerDatabaseTrigger(functionID, collection, event)
			}
		}
	}

	// Register cron triggers
	if cronTriggers, ok := triggerConfig["cron"].([]interface{}); ok {
		for _, trigger := range cronTriggers {
			if cronTrigger, ok := trigger.(map[string]interface{}); ok {
				schedule := cronTrigger["schedule"].(string)
				p.registerCronTrigger(functionID, schedule)
			}
		}
	}

	return nil
}

// registerHTTPTrigger registers an HTTP trigger for an lambda function
func (p *LambdaFunctionPlugin) registerHTTPTrigger(functionID, method, path string) {
	routeKey := fmt.Sprintf("%s:%s", method, path)
	route := &LambdaFunctionHTTPRoute{
		FunctionID: functionID,
		Method:     method,
		Path:       path,
		Handler:    p.createHTTPHandler(functionID),
	}
	p.httpRoutes.Store(routeKey, route)
}

// registerDatabaseTrigger registers a database trigger for an lambda function
func (p *LambdaFunctionPlugin) registerDatabaseTrigger(functionID, collection, event string) {
	trigger := &LambdaFunctionDBTrigger{
		FunctionID: functionID,
		Collection: collection,
		Event:      event,
	}

	key := fmt.Sprintf("%s:%s", collection, event)
	triggers, _ := p.dbTriggers.LoadOrStore(key, []*LambdaFunctionDBTrigger{})
	updatedTriggers := append(triggers.([]*LambdaFunctionDBTrigger), trigger)
	p.dbTriggers.Store(key, updatedTriggers)
}

// registerCronTrigger registers a cron trigger for an lambda function
func (p *LambdaFunctionPlugin) registerCronTrigger(functionID, schedule string) {
	jobID := fmt.Sprintf("lambda_function_%s", functionID)
	job := &LambdaFunctionCronJob{
		FunctionID: functionID,
		Schedule:   schedule,
		JobID:      jobID,
	}

	p.scheduler.MustAdd(jobID, schedule, func() {
		p.executeFunctionForCron(functionID)
	})

	p.cronJobs.Store(functionID, job)
}

// registerHTTPRoutes registers HTTP routes with the PocketBase router
func (p *LambdaFunctionPlugin) registerHTTPRoutes() {
	if p.router == nil {
		p.app.Logger().Debug("Router not available yet, skipping HTTP route registration")
		return
	}
	
	p.httpRoutes.Range(func(key, value interface{}) bool {
		route := value.(*LambdaFunctionHTTPRoute)
		
		// Support both prefixed and direct routes
		// If path starts with /api/, use as-is
		// Otherwise, use direct path for custom routes like /test, /ui
		var fullPath string
		if strings.HasPrefix(route.Path, "/api/") {
			fullPath = route.Path
		} else {
			// Custom routes without prefix
			fullPath = route.Path
		}
		
		p.app.Logger().Info("Registering lambda HTTP route", 
			"method", route.Method, 
			"path", fullPath, 
			"function", route.FunctionID)
		
		p.router.Route(route.Method, fullPath, route.Handler)
		return true
	})
}

// registerDatabaseTriggers registers database event triggers
func (p *LambdaFunctionPlugin) registerDatabaseTriggers() {
	// Register for record creation
	p.app.OnRecordCreate().BindFunc(func(e *core.RecordEvent) error {
		if err := e.Next(); err != nil {
			return err
		}
		return p.executeFunctionForDBEvent(e.Record, nil, "create")
	})

	// Register for record updates
	p.app.OnRecordUpdate().BindFunc(func(e *core.RecordEvent) error {
		oldRecord := e.Record.Original()
		if err := e.Next(); err != nil {
			return err
		}
		return p.executeFunctionForDBEvent(e.Record, oldRecord, "update")
	})

	// Register for record deletion
	p.app.OnRecordDelete().BindFunc(func(e *core.RecordEvent) error {
		if err := e.Next(); err != nil {
			return err
		}
		return p.executeFunctionForDBEvent(e.Record, nil, "delete")
	})
}

// createHTTPHandler creates an HTTP handler for an lambda function
func (p *LambdaFunctionPlugin) createHTTPHandler(functionID string) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		ctx := &LambdaFunctionExecutionContext{
			FunctionID:  functionID,
			TriggerType: "http",
			Request:     e.Request,
			Response:    e.Response,
			StartTime:   time.Now(),
		}

		result := p.executeFunction(ctx)
		
		if !result.Success {
			return e.InternalServerError("Lambda function execution failed", fmt.Errorf(result.Error))
		}

		p.app.Logger().Info("Lambda function result", 
			"type", fmt.Sprintf("%T", result.Output), 
			"value", fmt.Sprintf("%+v", result.Output))

		// Convert goja.Object to Go map
		var responseMap map[string]interface{}
		
		// If it's a goja.Object, convert it to a map
		if gojaObj, ok := result.Output.(*goja.Object); ok {
			if exported := gojaObj.Export(); exported != nil {
				if convertedMap, ok := exported.(map[string]interface{}); ok {
					responseMap = convertedMap
					p.app.Logger().Info("Converted goja.Object to map", "map", responseMap)
				}
			}
		} else if directMap, ok := result.Output.(map[string]interface{}); ok {
			responseMap = directMap
		}

		// If the function returned a response object, handle it
		if responseMap != nil {
			p.app.Logger().Info("Processing response object", "response", responseMap)
			// Set status code
			status := http.StatusOK
			if statusValue, ok := responseMap["status"].(float64); ok {
				status = int(statusValue)
			}
			
			// Set headers
			if headers, ok := responseMap["headers"].(map[string]interface{}); ok {
				for key, value := range headers {
					e.Response.Header().Set(key, fmt.Sprintf("%v", value))
				}
			}
			
			// Handle response body
			if body, ok := responseMap["body"]; ok {
				if bodyStr, ok := body.(string); ok {
					// Check if Content-Type header is set to determine response type
					contentType := e.Response.Header().Get("Content-Type")
					if contentType == "" {
						// Use function's configured content type as default
						functionRecord, err := p.app.FindRecordById(core.CollectionNameLambdaFunctions, functionID)
						if err == nil {
							configuredContentType := functionRecord.GetString("contentType")
							if configuredContentType != "" && configuredContentType != "auto" {
								contentType = configuredContentType
							} else {
								// Intelligent content type detection for "auto" mode
								contentType = p.detectContentType(bodyStr)
							}
						} else {
							// Fallback to text/plain if we can't find the function record
							contentType = "text/plain"
						}
						e.Response.Header().Set("Content-Type", contentType)
					}
					
					e.Response.WriteHeader(status)
					e.Response.Write([]byte(bodyStr))
					return nil
				}
				// Non-string body, return as JSON
				return e.JSON(status, body)
			}
			
			// No body, just return status
			e.Response.WriteHeader(status)
			return nil
		}

		// Default: return function output as JSON
		p.app.Logger().Info("Using default JSON response")
		return e.JSON(http.StatusOK, result.Output)
	}
}

// detectContentType intelligently detects content type based on content
func (p *LambdaFunctionPlugin) detectContentType(content string) string {
	content = strings.TrimSpace(content)
	
	// Check for HTML
	if strings.HasPrefix(content, "<!DOCTYPE html") || 
	   strings.HasPrefix(content, "<html") || 
	   strings.Contains(content, "<body") ||
	   strings.Contains(content, "<div") ||
	   strings.Contains(content, "<span") {
		return "text/html"
	}
	
	// Check for JSON
	if (strings.HasPrefix(content, "{") && strings.HasSuffix(content, "}")) ||
	   (strings.HasPrefix(content, "[") && strings.HasSuffix(content, "]")) {
		return "application/json"
	}
	
	// Check for XML
	if strings.HasPrefix(content, "<?xml") || 
	   (strings.HasPrefix(content, "<") && strings.Contains(content, ">")) {
		return "application/xml"
	}
	
	// Check for CSS
	if strings.Contains(content, "{") && strings.Contains(content, "}") && 
	   (strings.Contains(content, "color:") || strings.Contains(content, "font-") || 
	    strings.Contains(content, "margin:") || strings.Contains(content, "padding:")) {
		return "text/css"
	}
	
	// Check for JavaScript
	if strings.Contains(content, "function") || strings.Contains(content, "var ") ||
	   strings.Contains(content, "let ") || strings.Contains(content, "const ") ||
	   strings.Contains(content, "console.log") || strings.Contains(content, "document.") {
		return "application/javascript"
	}
	
	// Default to plain text
	return "text/plain"
}

// executeFunctionForDBEvent executes functions triggered by database events
func (p *LambdaFunctionPlugin) executeFunctionForDBEvent(record, oldRecord *core.Record, event string) error {
	collection := record.Collection().Name
	key := fmt.Sprintf("%s:%s", collection, event)

	if triggers, ok := p.dbTriggers.Load(key); ok {
		for _, trigger := range triggers.([]*LambdaFunctionDBTrigger) {
			ctx := &LambdaFunctionExecutionContext{
				FunctionID:  trigger.FunctionID,
				TriggerType: "database",
				Record:      record,
				OldRecord:   oldRecord,
				StartTime:   time.Now(),
			}

			// Execute async to not block database operations
			go func(ctx *LambdaFunctionExecutionContext) {
				result := p.executeFunction(ctx)
				if !result.Success {
					p.app.Logger().Error("Lambda function execution failed", 
						"function", ctx.FunctionID, 
						"error", result.Error)
				}
			}(ctx)
		}
	}

	return nil
}

// executeFunctionForCron executes functions triggered by cron
func (p *LambdaFunctionPlugin) executeFunctionForCron(functionID string) {
	ctx := &LambdaFunctionExecutionContext{
		FunctionID:  functionID,
		TriggerType: "cron",
		StartTime:   time.Now(),
	}

	result := p.executeFunction(ctx)
	if !result.Success {
		p.app.Logger().Error("Lambda function cron execution failed", 
			"function", functionID, 
			"error", result.Error)
	}
}

// executeFunction executes an lambda function with the given context
func (p *LambdaFunctionPlugin) executeFunction(ctx *LambdaFunctionExecutionContext) *LambdaFunctionExecutionResult {
	// Load function from database
	function, err := p.app.FindRecordById("lambdas", ctx.FunctionID)
	if err != nil {
		return &LambdaFunctionExecutionResult{
			Success:  false,
			Error:    fmt.Sprintf("Function not found: %v", err),
			Duration: time.Since(ctx.StartTime),
		}
	}

	if !function.GetBool("enabled") {
		return &LambdaFunctionExecutionResult{
			Success:  false,
			Error:    "Function is disabled",
			Duration: time.Since(ctx.StartTime),
		}
	}

	var result *LambdaFunctionExecutionResult

	// Execute with a fresh VM for true isolation
	// Instead of using the pool (which reuses VMs), create a fresh VM for each execution
	vm := p.createVM()
	
	// Set execution context
	p.setExecutionContext(vm, ctx, function)

	// Execute with timeout
	execCtx, cancel := context.WithTimeout(context.Background(), p.config.MaxExecutionTime)
	defer cancel()

	// Execute the function
	output, err := p.executeWithContext(execCtx, vm, function.GetString("code"))
	
	result = &LambdaFunctionExecutionResult{
		Success:  err == nil,
		Output:   output,
		Error:    p.formatError(err),
		Duration: time.Since(ctx.StartTime),
	}

	return result
}

// setExecutionContext sets the execution context in the VM
func (p *LambdaFunctionPlugin) setExecutionContext(vm *goja.Runtime, ctx *LambdaFunctionExecutionContext, function *core.Record) {
	// Set environment variables
	env := make(map[string]string)
	if envVars := function.GetString("env_vars"); envVars != "" {
		json.Unmarshal([]byte(envVars), &env)
	}
	vm.Set("$env", env)

	// Set trigger context
	vm.Set("$trigger", map[string]interface{}{
		"type":       ctx.TriggerType,
		"function":   function.GetString("name"),
		"timestamp":  ctx.StartTime.Unix(),
	})

	// Set request context for HTTP triggers
	if ctx.Request != nil {
		vm.Set("$request", map[string]interface{}{
			"method":  ctx.Request.Method,
			"url":     ctx.Request.URL.String(),
			"headers": ctx.Request.Header,
			"body":    p.getRequestBody(ctx.Request),
		})
	}

	// Set record context for database triggers
	if ctx.Record != nil {
		vm.Set("$record", ctx.Record)
		if ctx.OldRecord != nil {
			vm.Set("$oldRecord", ctx.OldRecord)
		}
	}
}

// executeWithContext executes JavaScript code with timeout
func (p *LambdaFunctionPlugin) executeWithContext(ctx context.Context, vm *goja.Runtime, code string) (interface{}, error) {
	done := make(chan struct{})
	var result interface{}
	var err error

	go func() {
		defer close(done)
		
		// Execute the code directly in a fresh VM
		// Each execution gets a completely isolated environment
		result, err = vm.RunString(code)
	}()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("execution timeout")
	case <-done:
		return result, err
	}
}

// getRequestBody extracts request body as string
func (p *LambdaFunctionPlugin) getRequestBody(r *http.Request) string {
	if r.Body == nil {
		return ""
	}
	
	body := make([]byte, 0, 1024)
	if _, err := r.Body.Read(body); err != nil {
		return ""
	}
	
	return string(body)
}

// formatError formats an error for output
func (p *LambdaFunctionPlugin) formatError(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// clearUserVariables clears user-defined variables from the VM while preserving PocketBase bindings
func (p *LambdaFunctionPlugin) clearUserVariables(vm *goja.Runtime) {
	// Instead of trying to selectively clear variables, which is complex and error-prone,
	// let's use a more direct approach: run code to delete user-defined variables
	
	// Get list of all current global properties
	_, err := vm.RunString(`
		(function() {
			// List of system properties to preserve
			const preserve = new Set([
				'$app', '$template', 'console', 'require', 'process', 'Buffer', 'global',
				'Object', 'Array', 'String', 'Number', 'Boolean', 'Date', 'Math', 'JSON',
				'RegExp', 'Error', 'TypeError', 'ReferenceError', 'SyntaxError', 'RangeError',
				'parseInt', 'parseFloat', 'isNaN', 'isFinite', 'encodeURI', 'decodeURI',
				'encodeURIComponent', 'decodeURIComponent', 'escape', 'unescape', 'eval',
				'undefined', 'NaN', 'Infinity', 'setTimeout', 'clearTimeout', 'setInterval', 'clearInterval'
			]);
			
			// Get all property names from global object
			const globalObj = (function() { return this; })();
			const allKeys = Object.getOwnPropertyNames(globalObj);
			
			// Delete user-defined properties
			for (const key of allKeys) {
				if (!preserve.has(key) && !key.startsWith('$')) {
					try {
						delete globalObj[key];
					} catch (e) {
						// Some properties might not be deletable, ignore errors
					}
				}
			}
		})();
	`)
	
	if err != nil {
		// If clearing fails, log it but don't fail the execution
		// This is a best-effort cleanup
	}
}

// Function lifecycle handlers
func (p *LambdaFunctionPlugin) handleFunctionCreated(record *core.Record) error {
	if err := p.registerFunction(record); err != nil {
		return err
	}
	// Re-register HTTP routes to include new function routes
	p.registerHTTPRoutes()
	return nil
}

func (p *LambdaFunctionPlugin) handleFunctionUpdated(record *core.Record) error {
	// Remove old registrations
	p.handleFunctionDeleted(record)
	// Register new ones
	if err := p.registerFunction(record); err != nil {
		return err
	}
	// Re-register HTTP routes to include updated function routes
	p.registerHTTPRoutes()
	return nil
}

func (p *LambdaFunctionPlugin) handleFunctionDeleted(record *core.Record) error {
	functionID := record.Id

	// Remove HTTP routes
	p.httpRoutes.Range(func(key, value interface{}) bool {
		route := value.(*LambdaFunctionHTTPRoute)
		if route.FunctionID == functionID {
			p.httpRoutes.Delete(key)
		}
		return true
	})

	// Remove database triggers
	p.dbTriggers.Range(func(key, value interface{}) bool {
		triggers := value.([]*LambdaFunctionDBTrigger)
		filtered := make([]*LambdaFunctionDBTrigger, 0)
		for _, trigger := range triggers {
			if trigger.FunctionID != functionID {
				filtered = append(filtered, trigger)
			}
		}
		if len(filtered) == 0 {
			p.dbTriggers.Delete(key)
		} else {
			p.dbTriggers.Store(key, filtered)
		}
		return true
	})

	// Remove cron jobs
	if job, ok := p.cronJobs.LoadAndDelete(functionID); ok {
		cronJob := job.(*LambdaFunctionCronJob)
		p.scheduler.Remove(cronJob.JobID)
	}

	return nil
}