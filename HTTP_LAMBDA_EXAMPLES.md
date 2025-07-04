# HTTP Lambda Function Examples

These examples show how to create Lambda functions that handle HTTP requests and return different content types.

## Basic HTTP Response

Simple JSON response:

```javascript
// Basic JSON response
{
  message: "Hello from Lambda!",
  timestamp: new Date().toISOString(),
  data: {
    version: "1.0.0",
    status: "active"
  }
}
```

## HTML Response

Return HTML content for web pages:

```javascript
// HTML response for /ui route
const htmlContent = `
<!DOCTYPE html>
<html>
<head>
    <title>Lambda UI</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            background: white;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        h1 {
            color: #333;
            border-bottom: 2px solid #007bff;
            padding-bottom: 10px;
        }
        .stats {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-top: 20px;
        }
        .stat-card {
            background: #f8f9fa;
            padding: 15px;
            border-radius: 5px;
            border-left: 4px solid #007bff;
        }
        .stat-number {
            font-size: 24px;
            font-weight: bold;
            color: #007bff;
        }
        .stat-label {
            color: #666;
            font-size: 14px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 Lambda Functions Dashboard</h1>
        <p>Welcome to your Lambda Functions control panel!</p>
        
        <div class="stats">
            <div class="stat-card">
                <div class="stat-number">12</div>
                <div class="stat-label">Active Functions</div>
            </div>
            <div class="stat-card">
                <div class="stat-number">1,234</div>
                <div class="stat-label">Executions Today</div>
            </div>
            <div class="stat-card">
                <div class="stat-number">98.5%</div>
                <div class="stat-label">Success Rate</div>
            </div>
            <div class="stat-card">
                <div class="stat-number">45ms</div>
                <div class="stat-label">Avg Response Time</div>
            </div>
        </div>
        
        <h2>Quick Actions</h2>
        <ul>
            <li><a href="/api/lambdas">View Lambda Functions API</a></li>
            <li><a href="/test">Test Lambda Route</a></li>
            <li><a href="/health">System Health Check</a></li>
        </ul>
        
        <p><small>Generated at ${new Date().toLocaleString()}</small></p>
    </div>
</body>
</html>
`;

// Return HTML response
({
  status: 200,
  headers: {
    "Content-Type": "text/html"
  },
  body: htmlContent
})
```

## Plain Text Response

Simple text response:

```javascript
// Plain text response
const message = `
Lambda Function Status Report
============================

Timestamp: ${new Date().toISOString()}
Server: PocketBase Lambda Functions
Status: Operational

System Information:
- Function: ${$trigger.function}
- Trigger: ${$trigger.type}
- Execution Time: ${Date.now() - $trigger.timestamp * 1000}ms

Request Details:
- Method: ${$request.method}
- URL: ${$request.url}
- Headers: ${Object.keys($request.headers).length} headers

End of Report
`;

({
  status: 200,
  headers: {
    "Content-Type": "text/plain"
  },
  body: message
})
```

## JSON API Response

Structured API response with custom headers:

```javascript
// JSON API response with custom headers
try {
  // Get some data
  const users = $app.findRecords("_pb_users_auth_", "", "created", 10);
  const lambdas = $app.findRecords("lambdas", "", "created", 10);
  
  const apiResponse = {
    success: true,
    data: {
      users: users.length,
      lambdas: lambdas.length,
      timestamp: new Date().toISOString()
    },
    meta: {
      version: "1.0.0",
      endpoint: "/api/test",
      method: $request.method
    }
  };
  
  console.log(`API call successful: ${users.length} users, ${lambdas.length} lambdas`);
  
  // Return JSON with custom headers
  ({
    status: 200,
    headers: {
      "Content-Type": "application/json",
      "X-API-Version": "1.0.0",
      "X-Response-Time": `${Date.now()}ms`,
      "Cache-Control": "no-cache"
    },
    body: apiResponse
  })
  
} catch (error) {
  console.log("API error:", error);
  
  ({
    status: 500,
    headers: {
      "Content-Type": "application/json"
    },
    body: {
      success: false,
      error: error.toString(),
      timestamp: new Date().toISOString()
    }
  })
}
```

## CSS Response

Serve CSS files:

```javascript
// CSS response for /styles.css
const cssContent = `
/* Lambda Functions Custom Styles */
body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
    line-height: 1.6;
    color: #333;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    margin: 0;
    padding: 0;
}

.lambda-container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 20px;
}

.lambda-header {
    background: rgba(255, 255, 255, 0.9);
    padding: 20px;
    border-radius: 10px;
    margin-bottom: 20px;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

.lambda-card {
    background: white;
    border-radius: 8px;
    padding: 20px;
    margin-bottom: 20px;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
    transition: transform 0.2s ease;
}

.lambda-card:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.15);
}

.lambda-button {
    background: #007bff;
    color: white;
    border: none;
    padding: 10px 20px;
    border-radius: 5px;
    cursor: pointer;
    font-size: 14px;
    transition: background-color 0.2s ease;
}

.lambda-button:hover {
    background: #0056b3;
}

.lambda-status {
    display: inline-block;
    padding: 4px 8px;
    border-radius: 12px;
    font-size: 12px;
    font-weight: bold;
    text-transform: uppercase;
}

.lambda-status.active {
    background: #d4edda;
    color: #155724;
}

.lambda-status.inactive {
    background: #f8d7da;
    color: #721c24;
}

@media (max-width: 768px) {
    .lambda-container {
        padding: 10px;
    }
    
    .lambda-header {
        padding: 15px;
    }
    
    .lambda-card {
        padding: 15px;
    }
}
`;

({
  status: 200,
  headers: {
    "Content-Type": "text/css",
    "Cache-Control": "public, max-age=3600"
  },
  body: cssContent
})
```

## JavaScript Response

Serve JavaScript files:

```javascript
// JavaScript response for /app.js
const jsContent = `
// Lambda Functions Dashboard JavaScript
console.log('Lambda Functions Dashboard loaded');

class LambdaDashboard {
    constructor() {
        this.init();
    }
    
    init() {
        this.loadStats();
        this.setupEventListeners();
    }
    
    async loadStats() {
        try {
            const response = await fetch('/api/stats');
            const data = await response.json();
            this.updateStats(data);
        } catch (error) {
            console.error('Failed to load stats:', error);
        }
    }
    
    updateStats(data) {
        const elements = {
            functions: document.querySelector('.stat-functions'),
            executions: document.querySelector('.stat-executions'),
            success: document.querySelector('.stat-success'),
            response: document.querySelector('.stat-response')
        };
        
        if (elements.functions) elements.functions.textContent = data.functions || '0';
        if (elements.executions) elements.executions.textContent = data.executions || '0';
        if (elements.success) elements.success.textContent = (data.success || 0) + '%';
        if (elements.response) elements.response.textContent = (data.avgResponse || 0) + 'ms';
    }
    
    setupEventListeners() {
        document.addEventListener('click', (e) => {
            if (e.target.classList.contains('lambda-refresh')) {
                this.loadStats();
            }
        });
    }
}

// Initialize dashboard when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    new LambdaDashboard();
});
`;

({
  status: 200,
  headers: {
    "Content-Type": "application/javascript",
    "Cache-Control": "public, max-age=3600"
  },
  body: jsContent
})
```

## XML Response

Return XML data:

```javascript
// XML response
const xmlContent = `<?xml version="1.0" encoding="UTF-8"?>
<LambdaResponse>
    <Status>success</Status>
    <Timestamp>${new Date().toISOString()}</Timestamp>
    <Data>
        <Function>${$trigger.function}</Function>
        <TriggerType>${$trigger.type}</TriggerType>
        <RequestMethod>${$request.method}</RequestMethod>
        <RequestURL>${$request.url}</RequestURL>
    </Data>
    <Message>Lambda function executed successfully</Message>
</LambdaResponse>`;

({
  status: 200,
  headers: {
    "Content-Type": "application/xml"
  },
  body: xmlContent
})
```

## Configuration Instructions

1. **Create a Lambda Function** in the admin UI
2. **Set HTTP Trigger** with:
   - Method: `GET` (or `POST`, `PUT`, `DELETE`)
   - Path: `/test`, `/ui`, `/styles.css`, `/app.js`, etc.
3. **Copy the code** from any example above
4. **Save and Enable** the function
5. **Test the route** by visiting the URL in your browser

## Route Examples

- `/ui` - HTML dashboard
- `/test` - JSON API endpoint
- `/health` - Plain text health check
- `/styles.css` - CSS stylesheet
- `/app.js` - JavaScript application
- `/api/custom` - Custom API endpoint

## Pro Tips

- 🎨 Use `text/html` for HTML pages
- 📊 Use `application/json` for API responses
- 📝 Use `text/plain` for simple text
- 🎭 Use `text/css` for stylesheets
- ⚡ Use `application/javascript` for JS files
- 📦 Use `application/xml` for XML data
- 🔒 Add `Cache-Control` headers for static content
- 📱 Include responsive design for mobile compatibility

Your Lambda functions can now serve complete web applications! 🚀