# Advanced Lambda Function Examples

## 7. Email Notification System

Trigger notifications based on data changes:

```javascript
// Email notification system
const notification = {
    timestamp: new Date().toISOString(),
    type: "user_activity",
    notifications: []
};

try {
    // Check for new users in the last hour
    const oneHourAgo = new Date();
    oneHourAgo.setHours(oneHourAgo.getHours() - 1);
    
    const allUsers = $app.findRecords("_pb_users_auth_", "", "-created", 50);
    const newUsers = allUsers.filter(user => {
        const userDate = new Date(user.created);
        return userDate > oneHourAgo;
    });
    
    if (newUsers.length > 0) {
        notification.notifications.push({
            type: "new_users",
            count: newUsers.length,
            message: `${newUsers.length} new users registered in the last hour`,
            users: newUsers.map(u => ({
                id: u.id,
                email: u.email,
                created: u.created
            }))
        });
        
        console.log(`Alert: ${newUsers.length} new users registered!`);
    }
    
    // Check lambda function errors
    const recentLogs = $app.findRecords("lambda_logs", "", "-created", 20);
    const errorLogs = recentLogs.filter(log => !log.success);
    
    if (errorLogs.length > 0) {
        notification.notifications.push({
            type: "lambda_errors",
            count: errorLogs.length,
            message: `${errorLogs.length} lambda function errors detected`,
            errors: errorLogs.map(log => ({
                function: log.function_name,
                error: log.error,
                timestamp: log.created
            }))
        });
        
        console.log(`Warning: ${errorLogs.length} lambda errors found!`);
    }
    
    // System health summary
    notification.summary = {
        newUsers: newUsers.length,
        totalErrors: errorLogs.length,
        alertLevel: errorLogs.length > 5 ? "high" : newUsers.length > 10 ? "medium" : "low"
    };
    
} catch (error) {
    notification.error = error.toString();
    console.log("Notification system error:", error);
}

notification
```

## 8. Data Validation and Cleanup

Validate and clean data across collections:

```javascript
// Data validation and cleanup
const validation = {
    timestamp: new Date().toISOString(),
    results: {},
    issues: [],
    summary: {}
};

try {
    // Validate user emails
    const users = $app.findRecords("_pb_users_auth_", "", "created", 100);
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    let invalidEmails = 0;
    let duplicateEmails = 0;
    const emailCounts = {};
    
    users.forEach(user => {
        // Check email format
        if (!emailRegex.test(user.email)) {
            invalidEmails++;
            validation.issues.push({
                type: "invalid_email",
                userId: user.id,
                email: user.email,
                message: "Invalid email format"
            });
        }
        
        // Check for duplicates
        if (emailCounts[user.email]) {
            duplicateEmails++;
            validation.issues.push({
                type: "duplicate_email",
                userId: user.id,
                email: user.email,
                message: "Duplicate email found"
            });
        } else {
            emailCounts[user.email] = 1;
        }
    });
    
    validation.results.users = {
        total: users.length,
        invalidEmails,
        duplicateEmails,
        validUsers: users.length - invalidEmails - duplicateEmails
    };
    
    // Validate lambda functions
    const lambdas = $app.findRecords("lambdas");
    let invalidLambdas = 0;
    
    lambdas.forEach(lambda => {
        if (!lambda.name || lambda.name.length < 3) {
            invalidLambdas++;
            validation.issues.push({
                type: "invalid_lambda_name",
                lambdaId: lambda.id,
                name: lambda.name,
                message: "Lambda name too short or missing"
            });
        }
        
        if (!lambda.code || lambda.code.trim().length < 10) {
            invalidLambdas++;
            validation.issues.push({
                type: "invalid_lambda_code",
                lambdaId: lambda.id,
                name: lambda.name,
                message: "Lambda code is too short or empty"
            });
        }
    });
    
    validation.results.lambdas = {
        total: lambdas.length,
        invalid: invalidLambdas,
        valid: lambdas.length - invalidLambdas
    };
    
    // Overall summary
    validation.summary = {
        totalIssues: validation.issues.length,
        criticalIssues: validation.issues.filter(i => i.type.includes("duplicate") || i.type.includes("invalid")).length,
        dataQuality: validation.issues.length === 0 ? "excellent" : validation.issues.length < 5 ? "good" : "needs_attention"
    };
    
    console.log(`Validation complete: ${validation.issues.length} issues found`);
    
} catch (error) {
    validation.error = error.toString();
    console.log("Validation failed:", error);
}

validation
```

## 9. Performance Monitor

Monitor system performance and usage patterns:

```javascript
// Performance monitoring
const performance = {
    timestamp: new Date().toISOString(),
    metrics: {},
    trends: {},
    recommendations: []
};

try {
    // Analyze lambda execution performance
    const logs = $app.findRecords("lambda_logs", "", "-created", 200);
    const now = new Date();
    
    // Performance buckets
    const timeBuckets = {
        last1Hour: logs.filter(log => new Date(log.created) > new Date(now.getTime() - 60*60*1000)),
        last6Hours: logs.filter(log => new Date(log.created) > new Date(now.getTime() - 6*60*60*1000)),
        last24Hours: logs.filter(log => new Date(log.created) > new Date(now.getTime() - 24*60*60*1000))
    };
    
    Object.keys(timeBuckets).forEach(period => {
        const periodLogs = timeBuckets[period];
        const durations = periodLogs.map(log => log.duration_ms || 0);
        const successes = periodLogs.filter(log => log.success).length;
        
        performance.metrics[period] = {
            executions: periodLogs.length,
            successRate: periodLogs.length > 0 ? Math.round((successes / periodLogs.length) * 100) : 0,
            avgDuration: durations.length > 0 ? Math.round(durations.reduce((a, b) => a + b, 0) / durations.length) : 0,
            maxDuration: durations.length > 0 ? Math.max(...durations) : 0,
            minDuration: durations.length > 0 ? Math.min(...durations) : 0
        };
    });
    
    // Function-specific performance
    const functionPerf = {};
    logs.forEach(log => {
        if (!functionPerf[log.function_name]) {
            functionPerf[log.function_name] = {
                executions: 0,
                totalDuration: 0,
                successes: 0,
                failures: 0
            };
        }
        
        const func = functionPerf[log.function_name];
        func.executions++;
        func.totalDuration += log.duration_ms || 0;
        
        if (log.success) {
            func.successes++;
        } else {
            func.failures++;
        }
    });
    
    // Calculate averages and add recommendations
    Object.keys(functionPerf).forEach(funcName => {
        const func = functionPerf[funcName];
        func.avgDuration = Math.round(func.totalDuration / func.executions);
        func.successRate = Math.round((func.successes / func.executions) * 100);
        
        // Add recommendations
        if (func.avgDuration > 5000) {
            performance.recommendations.push({
                type: "performance",
                function: funcName,
                issue: "Slow execution",
                suggestion: `Function ${funcName} averages ${func.avgDuration}ms - consider optimization`
            });
        }
        
        if (func.successRate < 90) {
            performance.recommendations.push({
                type: "reliability",
                function: funcName,
                issue: "Low success rate",
                suggestion: `Function ${funcName} has ${func.successRate}% success rate - needs debugging`
            });
        }
    });
    
    performance.trends.functionPerformance = functionPerf;
    
    // System load estimation
    const totalExecutionsLast24h = performance.metrics.last24Hours?.executions || 0;
    performance.trends.systemLoad = {
        executionsPerHour: Math.round(totalExecutionsLast24h / 24),
        executionsPerMinute: Math.round(totalExecutionsLast24h / (24 * 60)),
        loadLevel: totalExecutionsLast24h > 1000 ? "high" : totalExecutionsLast24h > 100 ? "medium" : "low"
    };
    
    console.log(`Performance analysis: ${totalExecutionsLast24h} executions in 24h, ${performance.recommendations.length} recommendations`);
    
} catch (error) {
    performance.error = error.toString();
    console.log("Performance monitoring failed:", error);
}

performance
```

## 10. Automated Testing Function

Test other lambda functions automatically:

```javascript
// Automated testing for lambda functions
const testing = {
    timestamp: new Date().toISOString(),
    tests: [],
    results: {},
    summary: {}
};

try {
    // Get all enabled lambda functions
    const lambdas = $app.findRecords("lambdas", "enabled = true");
    
    let totalTests = 0;
    let passedTests = 0;
    let failedTests = 0;
    
    lambdas.forEach(lambda => {
        const test = {
            functionName: lambda.name,
            functionId: lambda.id,
            tests: []
        };
        
        // Test 1: Code syntax check (basic)
        const codeTest = {
            name: "code_syntax",
            description: "Check if code contains basic JavaScript syntax",
            passed: false,
            message: ""
        };
        
        if (lambda.code && lambda.code.includes("console.log")) {
            codeTest.passed = true;
            codeTest.message = "Code contains console.log statements";
            passedTests++;
        } else {
            codeTest.message = "Code missing console.log or basic JS structure";
            failedTests++;
        }
        test.tests.push(codeTest);
        totalTests++;
        
        // Test 2: Timeout validation
        const timeoutTest = {
            name: "timeout_validation",
            description: "Check if timeout is reasonable",
            passed: false,
            message: ""
        };
        
        if (lambda.timeout >= 1000 && lambda.timeout <= 300000) {
            timeoutTest.passed = true;
            timeoutTest.message = `Timeout ${lambda.timeout}ms is within acceptable range`;
            passedTests++;
        } else {
            timeoutTest.message = `Timeout ${lambda.timeout}ms is outside recommended range (1-300 seconds)`;
            failedTests++;
        }
        test.tests.push(timeoutTest);
        totalTests++;
        
        // Test 3: Recent execution check
        const executionTest = {
            name: "recent_execution",
            description: "Check if function has been executed recently",
            passed: false,
            message: ""
        };
        
        const recentLogs = $app.findRecords("lambda_logs", `function_id = "${lambda.id}"`, "-created", 5);
        if (recentLogs.length > 0) {
            const lastExecution = new Date(recentLogs[0].created);
            const hoursSinceExecution = (new Date() - lastExecution) / (1000 * 60 * 60);
            
            if (hoursSinceExecution < 24) {
                executionTest.passed = true;
                executionTest.message = `Last executed ${Math.round(hoursSinceExecution)} hours ago`;
                passedTests++;
            } else {
                executionTest.message = `Not executed in ${Math.round(hoursSinceExecution)} hours`;
                failedTests++;
            }
        } else {
            executionTest.message = "No execution history found";
            failedTests++;
        }
        test.tests.push(executionTest);
        totalTests++;
        
        // Test 4: Name convention check
        const nameTest = {
            name: "naming_convention",
            description: "Check if function name follows conventions",
            passed: false,
            message: ""
        };
        
        if (lambda.name && lambda.name.length >= 3 && /^[a-zA-Z][a-zA-Z0-9_-]*$/.test(lambda.name)) {
            nameTest.passed = true;
            nameTest.message = "Function name follows naming conventions";
            passedTests++;
        } else {
            nameTest.message = "Function name doesn't follow conventions (should start with letter, use only alphanumeric, _, -)";
            failedTests++;
        }
        test.tests.push(nameTest);
        totalTests++;
        
        testing.tests.push(test);
    });
    
    // Calculate overall results
    testing.results = {
        totalFunctions: lambdas.length,
        totalTests,
        passedTests,
        failedTests,
        passRate: totalTests > 0 ? Math.round((passedTests / totalTests) * 100) : 0
    };
    
    testing.summary = {
        status: testing.results.passRate >= 80 ? "good" : testing.results.passRate >= 60 ? "warning" : "critical",
        recommendation: testing.results.passRate >= 80 ? 
            "All functions are well-maintained" : 
            `${failedTests} tests failed - review function configurations`,
        nextAction: testing.results.failedTests > 0 ? 
            "Review failed tests and fix issues" : 
            "Consider adding more comprehensive tests"
    };
    
    console.log(`Testing complete: ${passedTests}/${totalTests} tests passed (${testing.results.passRate}%)`);
    
} catch (error) {
    testing.error = error.toString();
    console.log("Testing failed:", error);
}

testing
```

## How to Use These Advanced Functions

1. **Performance Monitor**: Set as a cron job to run every hour
2. **Email Notifications**: Trigger on database events or run periodically
3. **Data Validation**: Run weekly for data quality checks
4. **Automated Testing**: Schedule daily to ensure function health

## Pro Tips

- 🔧 **Combine functions**: Use one function to call others for complex workflows
- 📊 **Add logging**: All functions include comprehensive console.log statements
- ⚡ **Optimize performance**: Monitor execution times and optimize slow functions
- 🛡️ **Error handling**: Every function includes try-catch blocks
- 📈 **Track metrics**: Use the analytics functions to monitor your system

These functions showcase the full power of your Lambda Functions system! 🚀