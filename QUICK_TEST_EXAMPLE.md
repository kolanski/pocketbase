# Quick Test Example

Here's a simple example to test the HTTP lambda functionality:

## Test Function: Simple HTML Page

Create a lambda function with this code:

```javascript
// Simple HTML response for testing
const html = `
<!DOCTYPE html>
<html>
<head>
    <title>Lambda Test</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            max-width: 600px;
            margin: 50px auto;
            padding: 20px;
            background: #f0f0f0;
        }
        .container {
            background: white;
            padding: 30px;
            border-radius: 10px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        h1 {
            color: #333;
            text-align: center;
        }
        .info {
            background: #e3f2fd;
            padding: 15px;
            border-radius: 5px;
            margin: 20px 0;
        }
        .success {
            color: #2e7d32;
            font-weight: bold;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🎉 Lambda Function Success!</h1>
        <div class="info">
            <p><strong>Function:</strong> ${$trigger.function}</p>
            <p><strong>Trigger:</strong> ${$trigger.type}</p>
            <p><strong>Time:</strong> ${new Date().toLocaleString()}</p>
            <p><strong>URL:</strong> ${$request.url}</p>
            <p><strong>Method:</strong> ${$request.method}</p>
        </div>
        <p class="success">✅ HTTP triggers are working correctly!</p>
        <p>You can now create lambda functions that serve:</p>
        <ul>
            <li>HTML pages</li>
            <li>JSON APIs</li>
            <li>CSS files</li>
            <li>JavaScript files</li>
            <li>Plain text</li>
            <li>XML data</li>
        </ul>
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
  body: html
})
```

## Setup Instructions

1. **Create New Lambda Function**:
   - Name: `test-html-page`
   - Description: `Test HTML page for HTTP triggers`
   - Timeout: `30000` (30 seconds)

2. **Configure HTTP Trigger**:
   - Method: `GET`
   - Path: `/test`

3. **Paste the code above** into the code editor

4. **Save and Enable** the function

5. **Test the route**:
   - Visit: `http://localhost:8090/test`
   - You should see a styled HTML page

## Alternative Test: JSON API

For a simple JSON API test:

```javascript
// Simple JSON API response
const data = {
  success: true,
  message: "Lambda function is working!",
  timestamp: new Date().toISOString(),
  request: {
    method: $request.method,
    url: $request.url,
    headers: Object.keys($request.headers).length
  },
  function: {
    name: $trigger.function,
    type: $trigger.type
  }
};

console.log("API endpoint called successfully");

// Return JSON response
({
  status: 200,
  headers: {
    "Content-Type": "application/json"
  },
  body: data
})
```

Use `/api/test` as the path for this one.

## What's Fixed

✅ **HTTP routes now work**: Routes are registered after functions are loaded  
✅ **Custom content types**: Support for HTML, CSS, JS, XML, plain text  
✅ **Direct routes**: Routes like `/test` work without `/api/` prefix  
✅ **Response headers**: Set custom headers for caching, content type, etc.  
✅ **Status codes**: Return custom HTTP status codes  
✅ **Real-time registration**: New functions create routes immediately  

Your lambda functions can now serve complete web applications! 🚀