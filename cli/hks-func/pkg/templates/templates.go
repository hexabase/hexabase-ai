package templates

import (
	"fmt"
)

// Template represents a function template
type Template struct {
	Runtime   string
	Type      string
	Files     map[string]string
	GitIgnore string
}

// GetTemplate returns a template for the given runtime and type
func GetTemplate(runtime, templateType string) (*Template, error) {
	key := fmt.Sprintf("%s-%s", runtime, templateType)
	
	tmpl, exists := templates[key]
	if !exists {
		return nil, fmt.Errorf("template not found: %s", key)
	}
	
	return &tmpl, nil
}

// templates contains all available function templates
var templates = map[string]Template{
	// Node.js HTTP template
	"node-http": {
		Runtime: "node",
		Type:    "http",
		Files: map[string]string{
			"index.js": `const express = require('express');
const app = express();

// Middleware
app.use(express.json());
app.use(express.urlencoded({ extended: true }));

// Health check endpoint
app.get('/health', (req, res) => {
  res.json({ status: 'healthy' });
});

// Main handler
app.all('*', async (req, res) => {
  console.log('Request:', {
    method: req.method,
    path: req.path,
    headers: req.headers,
    body: req.body
  });

  // Your function logic here
  const response = {
    message: 'Hello from {{.FunctionName}}!',
    timestamp: new Date().toISOString(),
    request: {
      method: req.method,
      path: req.path
    }
  };

  res.json(response);
});

// Export handler for Knative
const port = process.env.PORT || 8080;
app.listen(port, () => {
  console.log('Function listening on port', port);
});

module.exports = { app };
`,
			"package.json": `{
  "name": "{{.FunctionName}}",
  "version": "1.0.0",
  "description": "{{.Description}}",
  "main": "index.js",
  "scripts": {
    "start": "node index.js",
    "test": "jest",
    "dev": "nodemon index.js"
  },
  "dependencies": {
    "express": "^4.18.2"
  },
  "devDependencies": {
    "jest": "^29.5.0",
    "nodemon": "^3.0.1",
    "supertest": "^6.3.3"
  }
}
`,
			"test/index.test.js": `const request = require('supertest');
const { app } = require('../index');

describe('Function tests', () => {
  test('Health check returns 200', async () => {
    const response = await request(app)
      .get('/health')
      .expect(200);
    
    expect(response.body).toHaveProperty('status', 'healthy');
  });

  test('Main handler returns expected response', async () => {
    const response = await request(app)
      .post('/')
      .send({ test: 'data' })
      .expect(200);
    
    expect(response.body).toHaveProperty('message');
    expect(response.body).toHaveProperty('timestamp');
  });
});
`,
		},
		GitIgnore: `node_modules/
.env
.env.local
*.log
coverage/
.DS_Store
`,
	},

	// Python HTTP template
	"python-http": {
		Runtime: "python",
		Type:    "http",
		Files: map[string]string{
			"main.py": `import os
import json
import logging
from datetime import datetime
from flask import Flask, request, jsonify

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Create Flask app
app = Flask(__name__)

@app.route('/health', methods=['GET'])
def health():
    """Health check endpoint"""
    return jsonify({'status': 'healthy'})

@app.route('/', defaults={'path': ''}, methods=['GET', 'POST', 'PUT', 'DELETE', 'PATCH'])
@app.route('/<path:path>', methods=['GET', 'POST', 'PUT', 'DELETE', 'PATCH'])
def handler(path):
    """Main function handler"""
    logger.info(f"Request: {request.method} {request.path}")
    logger.info(f"Headers: {dict(request.headers)}")
    
    # Get request data
    data = None
    if request.is_json:
        data = request.get_json()
    elif request.form:
        data = dict(request.form)
    
    # Your function logic here
    response = {
        'message': 'Hello from {{.FunctionName}}!',
        'timestamp': datetime.utcnow().isoformat(),
        'request': {
            'method': request.method,
            'path': request.path,
            'data': data
        }
    }
    
    return jsonify(response)

if __name__ == '__main__':
    port = int(os.environ.get('PORT', 8080))
    app.run(host='0.0.0.0', port=port)
`,
			"requirements.txt": `Flask==3.0.0
gunicorn==21.2.0
`,
			"test_main.py": `import pytest
import json
from main import app

@pytest.fixture
def client():
    app.config['TESTING'] = True
    with app.test_client() as client:
        yield client

def test_health_check(client):
    """Test health check endpoint"""
    response = client.get('/health')
    assert response.status_code == 200
    data = json.loads(response.data)
    assert data['status'] == 'healthy'

def test_main_handler(client):
    """Test main handler"""
    response = client.post('/', 
        data=json.dumps({'test': 'data'}),
        content_type='application/json'
    )
    assert response.status_code == 200
    data = json.loads(response.data)
    assert 'message' in data
    assert 'timestamp' in data
    assert data['request']['method'] == 'POST'
`,
			"Procfile": `web: gunicorn main:app --bind 0.0.0.0:$PORT
`,
		},
		GitIgnore: `__pycache__/
*.py[cod]
*$py.class
.env
.venv/
venv/
.pytest_cache/
.coverage
*.log
.DS_Store
`,
	},

	// Go HTTP template
	"go-http": {
		Runtime: "go",
		Type:    "http",
		Files: map[string]string{
			"main.go": `package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type Response struct {
	Message   string    ` + "`json:\"message\"`" + `
	Timestamp time.Time ` + "`json:\"timestamp\"`" + `
	Request   Request   ` + "`json:\"request\"`" + `
}

type Request struct {
	Method string      ` + "`json:\"method\"`" + `
	Path   string      ` + "`json:\"path\"`" + `
	Data   interface{} ` + "`json:\"data,omitempty\"`" + `
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request: %s %s", r.Method, r.URL.Path)
	
	// Parse request body
	var data interface{}
	if r.Body != nil {
		defer r.Body.Close()
		json.NewDecoder(r.Body).Decode(&data)
	}
	
	// Your function logic here
	response := Response{
		Message:   "Hello from {{.FunctionName}}!",
		Timestamp: time.Now().UTC(),
		Request: Request{
			Method: r.Method,
			Path:   r.URL.Path,
			Data:   data,
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/", mainHandler)
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	log.Printf("Function listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
`,
			"go.mod": `module {{.FunctionName}}

go 1.21
`,
			"main_test.go": `package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(healthHandler)
	handler.ServeHTTP(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	
	var response map[string]string
	json.NewDecoder(rr.Body).Decode(&response)
	
	if response["status"] != "healthy" {
		t.Errorf("handler returned unexpected body: got %v want %v", response["status"], "healthy")
	}
}

func TestMainHandler(t *testing.T) {
	body := strings.NewReader(` + "`" + `{"test": "data"}` + "`" + `)
	req, err := http.NewRequest("POST", "/", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(mainHandler)
	handler.ServeHTTP(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	
	var response Response
	json.NewDecoder(rr.Body).Decode(&response)
	
	if response.Message == "" {
		t.Error("handler returned empty message")
	}
	
	if response.Request.Method != "POST" {
		t.Errorf("handler returned wrong method: got %v want %v", response.Request.Method, "POST")
	}
}
`,
		},
		GitIgnore: `*.exe
*.dll
*.so
*.dylib
*.test
*.out
.env
vendor/
.DS_Store
`,
	},

	// Python Event template
	"python-event": {
		Runtime: "python",
		Type:    "event",
		Files: map[string]string{
			"main.py": `import os
import json
import logging
from datetime import datetime
from cloudevents.http import from_http
from flask import Flask, request, jsonify

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Create Flask app
app = Flask(__name__)

@app.route('/health', methods=['GET'])
def health():
    """Health check endpoint"""
    return jsonify({'status': 'healthy'})

@app.route('/', methods=['POST'])
def handler():
    """CloudEvents handler"""
    # Parse CloudEvent
    event = from_http(request.headers, request.get_data())
    
    logger.info(f"Received event: {event['type']} from {event['source']}")
    logger.info(f"Event ID: {event['id']}")
    logger.info(f"Event data: {event.data}")
    
    # Your event processing logic here
    result = process_event(event)
    
    # Return response
    response = {
        'message': 'Event processed successfully',
        'timestamp': datetime.utcnow().isoformat(),
        'event_id': event['id'],
        'event_type': event['type'],
        'result': result
    }
    
    return jsonify(response)

def process_event(event):
    """Process the incoming event"""
    # Add your event processing logic here
    event_type = event['type']
    event_data = event.data
    
    # Example: Handle different event types
    if event_type == 'order.created':
        return handle_order_created(event_data)
    elif event_type == 'user.registered':
        return handle_user_registered(event_data)
    else:
        return {'status': 'processed', 'type': event_type}

def handle_order_created(data):
    """Handle order created event"""
    logger.info(f"Processing order: {data.get('order_id')}")
    # Add your order processing logic
    return {'order_id': data.get('order_id'), 'status': 'processed'}

def handle_user_registered(data):
    """Handle user registered event"""
    logger.info(f"Processing new user: {data.get('user_id')}")
    # Add your user processing logic
    return {'user_id': data.get('user_id'), 'status': 'welcomed'}

if __name__ == '__main__':
    port = int(os.environ.get('PORT', 8080))
    app.run(host='0.0.0.0', port=port)
`,
			"requirements.txt": `Flask==3.0.0
cloudevents==1.10.1
gunicorn==21.2.0
`,
			"test_main.py": `import pytest
import json
from main import app
from datetime import datetime

@pytest.fixture
def client():
    app.config['TESTING'] = True
    with app.test_client() as client:
        yield client

def test_health_check(client):
    """Test health check endpoint"""
    response = client.get('/health')
    assert response.status_code == 200
    data = json.loads(response.data)
    assert data['status'] == 'healthy'

def test_event_handler(client):
    """Test CloudEvent handler"""
    # Create test CloudEvent
    headers = {
        'ce-specversion': '1.0',
        'ce-type': 'order.created',
        'ce-source': 'test/orders',
        'ce-id': 'test-123',
        'ce-time': datetime.utcnow().isoformat() + 'Z',
        'content-type': 'application/json'
    }
    
    data = {
        'order_id': 'ORD-123',
        'amount': 99.99
    }
    
    response = client.post('/', 
        data=json.dumps(data),
        headers=headers
    )
    
    assert response.status_code == 200
    result = json.loads(response.data)
    assert result['message'] == 'Event processed successfully'
    assert result['event_id'] == 'test-123'
    assert result['event_type'] == 'order.created'
`,
		},
		GitIgnore: `__pycache__/
*.py[cod]
*$py.class
.env
.venv/
venv/
.pytest_cache/
.coverage
*.log
.DS_Store
`,
	},
}