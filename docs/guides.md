# Development Environment Setup

This guide will help you set up a complete development environment for Hexabase KaaS.

## Quick Setup (Automated)

We provide an automated setup script that handles most of the configuration for you:

```bash
# Clone the repository
git clone https://github.com/hexabase/hexabase-ai.git
cd hexabase-ai

# Run the setup script
./scripts/dev-setup.sh
```

The script will:
1. Check for required dependencies
2. Create a Kind cluster with ingress support
3. Install necessary Kubernetes components
4. Start PostgreSQL, Redis, and NATS using Docker Compose
5. Generate JWT keys
6. Create .env files with proper configuration
7. Update /etc/hosts for local access

After running the script, you can start developing:

```bash
# Terminal 1: Start the API
cd api
go run cmd/api/main.go

# Terminal 2: Start the UI
cd ui
npm install
npm run dev
```

Access the application at:
- API: http://api.localhost
- UI: http://app.localhost

## Manual Setup (Alternative)

If you prefer to set up the environment manually or need to understand what the script does, follow these steps:

### Prerequisites

### Required Software

1. **Go** (1.24 or later)
   ```bash
   # macOS
   brew install go
   
   # Linux
   wget https://go.dev/dl/go1.24.3.linux-amd64.tar.gz
   sudo tar -C /usr/local -xzf go1.24.3.linux-amd64.tar.gz
   export PATH=$PATH:/usr/local/go/bin
   ```

2. **Node.js** (18.x or later) and npm
   ```bash
   # macOS
   brew install node
   
   # Linux (using NodeSource)
   curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
   sudo apt-get install -y nodejs
   ```

3. **Docker** and Docker Compose
   ```bash
   # macOS
   brew install --cask docker
   
   # Linux
   curl -fsSL https://get.docker.com | sh
   sudo usermod -aG docker $USER
   ```

4. **Kubernetes Tools**
   ```bash
   # kubectl
   brew install kubectl
   
   # k3s (lightweight Kubernetes)
   curl -sfL https://get.k3s.io | sh -
   
   # OR kind (Kubernetes in Docker)
   brew install kind
   ```

5. **PostgreSQL Client**
   ```bash
   # macOS
   brew install postgresql
   
   # Linux
   sudo apt-get install postgresql-client
   ```

6. **Redis Client**
   ```bash
   # macOS
   brew install redis
   
   # Linux
   sudo apt-get install redis-tools
   ```

### Development Tools

1. **Wire** (Dependency Injection)
   ```bash
   go install github.com/google/wire/cmd/wire@latest
   ```

2. **golang-migrate** (Database Migrations)
   ```bash
   brew install golang-migrate
   # OR
   go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
   ```

3. **golangci-lint** (Linting)
   ```bash
   brew install golangci-lint
   # OR
   curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
   ```

4. **Playwright** (E2E Testing)
   ```bash
   cd ui
   npm install
   npx playwright install
   ```

## Setting Up the Development Environment

### 1. Clone the Repository

```bash
git clone https://github.com/hexabase/hexabase-ai.git
cd hexabase-ai
```

### 2. Start Infrastructure Services

Create a `docker-compose.override.yml` for local development:

```yaml
version: '3.8'

services:
  postgres:
    ports:
      - "5432:5432"
    environment:
      POSTGRES_PASSWORD: localdev

  redis:
    ports:
      - "6379:6379"

  nats:
    image: nats:latest
    ports:
      - "4222:4222"
      - "8222:8222"
    command: "-js"
```

Start the services:

```bash
docker-compose up -d postgres redis nats
```

### 3. Set Up the Database

```bash
# Create database
psql -h localhost -U postgres -c "CREATE DATABASE hexabase_kaas;"

# Run migrations
cd api
migrate -path ./migrations -database "postgresql://postgres:localdev@localhost:5432/hexabase_kaas?sslmode=disable" up
```

### 4. Configure Environment Variables

Create `.env` files for both API and UI:

**api/.env**
```env
# Database
DATABASE_URL=postgres://postgres:localdev@localhost:5432/hexabase_kaas?sslmode=disable

# Redis
REDIS_URL=redis://localhost:6379

# NATS
NATS_URL=nats://localhost:4222

# JWT Keys (generate with: openssl genrsa -out private.pem 2048)
JWT_PRIVATE_KEY_PATH=./keys/private.pem
JWT_PUBLIC_KEY_PATH=./keys/public.pem

# OAuth Providers (replace with your values)
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret

# Stripe (for billing)
STRIPE_API_KEY=sk_test_...
STRIPE_WEBHOOK_SECRET=whsec_...

# Kubernetes
KUBECONFIG=$HOME/.kube/config
```

**ui/.env.local**
```env
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_WS_URL=ws://localhost:8080
```

### 5. Generate JWT Keys

```bash
cd api
mkdir -p keys
openssl genrsa -out keys/private.pem 2048
openssl rsa -in keys/private.pem -pubout -out keys/public.pem
```

### 6. Set Up Local Kubernetes Cluster

Using k3s:
```bash
# Install k3s (if not already installed)
curl -sfL https://get.k3s.io | sh -

# Copy kubeconfig
mkdir -p ~/.kube
sudo cp /etc/rancher/k3s/k3s.yaml ~/.kube/config
sudo chown $USER:$USER ~/.kube/config
```

OR using kind:
```bash
# Create cluster
kind create cluster --name hexabase-dev

# Install vCluster operator
kubectl create namespace vcluster
kubectl apply -f https://github.com/loft-sh/vcluster/releases/latest/download/vcluster-k8s.yaml
```

### 7. Build and Run the API

```bash
cd api

# Download dependencies
go mod download

# Generate wire dependencies
wire ./internal/infrastructure/wire

# Run tests
go test ./...

# Run the API server
go run cmd/api/main.go
```

### 8. Build and Run the UI

```bash
cd ui

# Install dependencies
npm install

# Run development server
npm run dev
```

## Development Workflow

### Running Tests

**API Tests:**
```bash
cd api
go test ./...                    # Run all tests
go test ./... -cover            # With coverage
go test -v ./internal/auth/...  # Specific package
```

**UI Tests:**
```bash
cd ui
npm test                        # Unit tests
npm run test:e2e               # Playwright E2E tests
```

### Code Quality Checks

**API:**
```bash
# Linting
golangci-lint run

# Format code
go fmt ./...

# Vet code
go vet ./...
```

**UI:**
```bash
# Linting
npm run lint

# Type checking
npm run type-check

# Format code
npm run format
```

### Database Migrations

```bash
# Create a new migration
migrate create -ext sql -dir api/migrations -seq add_new_table

# Apply migrations
migrate -path api/migrations -database $DATABASE_URL up

# Rollback
migrate -path api/migrations -database $DATABASE_URL down 1
```

## IDE Setup

### VS Code

Install recommended extensions:
- Go (official)
- ESLint
- Prettier
- GitLens
- Docker
- Kubernetes

Create `.vscode/settings.json`:
```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "workspace",
  "editor.formatOnSave": true,
  "[go]": {
    "editor.defaultFormatter": "golang.go"
  },
  "[typescript]": {
    "editor.defaultFormatter": "esbenp.prettier-vscode"
  }
}
```

### JetBrains GoLand/WebStorm

- Enable Go modules support
- Configure JavaScript version to ES2022
- Set up file watchers for formatting

## Troubleshooting

### Common Issues

1. **Port already in use**
   ```bash
   # Find process using port
   lsof -i :8080
   # Kill process
   kill -9 <PID>
   ```

2. **Database connection errors**
   - Ensure PostgreSQL is running: `docker-compose ps`
   - Check credentials in `.env`
   - Verify database exists: `psql -h localhost -U postgres -l`

3. **Kubernetes connection issues**
   - Check cluster is running: `kubectl cluster-info`
   - Verify kubeconfig: `kubectl config current-context`

4. **Go module errors**
   ```bash
   go clean -modcache
   go mod download
   ```

## Next Steps

- Read the [API Development Guide](./api-development-guide.md)
- Explore the [Frontend Development Guide](./frontend-development-guide.md)
- Review [Code Style Guide](./code-style-guide.md)
- Understand the [Git Workflow](./git-workflow.md)# Frontend Development Guide

This guide covers best practices and patterns for developing the Hexabase AI frontend.

## Overview

The Hexabase AI frontend is built with:
- Next.js 14+ (App Router)
- TypeScript
- Tailwind CSS
- shadcn/ui components
- React Query for data fetching
- Zustand for state management

## Project Structure

```
ui/
├── src/
│   ├── app/              # Next.js app router pages
│   ├── components/       # React components
│   │   ├── ui/          # Base UI components (shadcn)
│   │   └── features/    # Feature-specific components
│   ├── hooks/           # Custom React hooks
│   ├── lib/             # Utilities and helpers
│   ├── services/        # API client services
│   ├── stores/          # Zustand stores
│   └── types/           # TypeScript types
├── public/              # Static assets
└── tests/              # Test files
```

## Development Patterns

### 1. Component Architecture

Follow a hierarchical component structure:

```typescript
// Base component (ui/)
export const Button = ({ children, variant, ...props }) => {
  return (
    <button className={cn(buttonVariants({ variant }))} {...props}>
      {children}
    </button>
  );
};

// Feature component (features/)
export const CreateWorkspaceButton = () => {
  const { mutate: createWorkspace } = useCreateWorkspace();
  
  return (
    <Button onClick={() => createWorkspace()}>
      Create Workspace
    </Button>
  );
};

// Page component (app/)
export default function WorkspacesPage() {
  return (
    <div>
      <h1>Workspaces</h1>
      <CreateWorkspaceButton />
      <WorkspaceList />
    </div>
  );
}
```

### 2. Data Fetching

Use React Query for server state management:

```typescript
// services/workspaces.ts
export const workspacesApi = {
  list: async (): Promise<Workspace[]> => {
    const response = await apiClient.get('/workspaces');
    return response.data;
  },
  
  create: async (data: CreateWorkspaceDto): Promise<Workspace> => {
    const response = await apiClient.post('/workspaces', data);
    return response.data;
  }
};

// hooks/use-workspaces.ts
export const useWorkspaces = () => {
  return useQuery({
    queryKey: ['workspaces'],
    queryFn: workspacesApi.list,
  });
};

export const useCreateWorkspace = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: workspacesApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['workspaces'] });
    },
  });
};
```

### 3. State Management

Use Zustand for client state:

```typescript
// stores/auth.ts
interface AuthState {
  user: User | null;
  token: string | null;
  login: (credentials: LoginCredentials) => Promise<void>;
  logout: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  token: null,
  
  login: async (credentials) => {
    const { user, token } = await authApi.login(credentials);
    set({ user, token });
    localStorage.setItem('token', token);
  },
  
  logout: () => {
    set({ user: null, token: null });
    localStorage.removeItem('token');
  },
}));
```

### 4. Type Safety

Define comprehensive types:

```typescript
// types/workspace.ts
export interface Workspace {
  id: string;
  name: string;
  description?: string;
  plan: 'starter' | 'pro' | 'enterprise';
  createdAt: string;
  updatedAt: string;
}

export interface CreateWorkspaceDto {
  name: string;
  description?: string;
  plan: Workspace['plan'];
}

// Use discriminated unions for complex states
export type WorkspaceState = 
  | { status: 'idle' }
  | { status: 'loading' }
  | { status: 'success'; data: Workspace[] }
  | { status: 'error'; error: Error };
```

## UI Components

### Using shadcn/ui

Install components as needed:

```bash
npx shadcn-ui@latest add button
npx shadcn-ui@latest add form
npx shadcn-ui@latest add dialog
```

Customize components:

```typescript
// components/ui/button.tsx
import { cn } from '@/lib/utils';

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'default' | 'destructive' | 'outline' | 'secondary' | 'ghost' | 'link';
  size?: 'default' | 'sm' | 'lg' | 'icon';
}

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, ...props }, ref) => {
    return (
      <button
        className={cn(buttonVariants({ variant, size, className }))}
        ref={ref}
        {...props}
      />
    );
  }
);
```

### Form Handling

Use React Hook Form with Zod validation:

```typescript
// components/features/create-workspace-form.tsx
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';

const createWorkspaceSchema = z.object({
  name: z.string().min(3).max(50),
  description: z.string().max(200).optional(),
  plan: z.enum(['starter', 'pro', 'enterprise']),
});

export const CreateWorkspaceForm = () => {
  const form = useForm<z.infer<typeof createWorkspaceSchema>>({
    resolver: zodResolver(createWorkspaceSchema),
    defaultValues: {
      name: '',
      plan: 'starter',
    },
  });
  
  const { mutate: createWorkspace } = useCreateWorkspace();
  
  const onSubmit = (data: z.infer<typeof createWorkspaceSchema>) => {
    createWorkspace(data);
  };
  
  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)}>
        <FormField
          control={form.control}
          name="name"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Name</FormLabel>
              <FormControl>
                <Input {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <Button type="submit">Create</Button>
      </form>
    </Form>
  );
};
```

## Styling

### Tailwind CSS

Use utility classes with consistent patterns:

```typescript
// Use cn() utility for conditional classes
import { cn } from '@/lib/utils';

export const Card = ({ className, active, ...props }) => {
  return (
    <div
      className={cn(
        'rounded-lg border bg-card p-6 shadow-sm',
        'hover:shadow-md transition-shadow',
        active && 'ring-2 ring-primary',
        className
      )}
      {...props}
    />
  );
};
```

### Design Tokens

Define consistent design tokens:

```css
/* globals.css */
@layer base {
  :root {
    --background: 0 0% 100%;
    --foreground: 222.2 84% 4.9%;
    --primary: 222.2 47.4% 11.2%;
    --primary-foreground: 210 40% 98%;
    --secondary: 210 40% 96.1%;
    --secondary-foreground: 222.2 47.4% 11.2%;
    --destructive: 0 84.2% 60.2%;
    --destructive-foreground: 210 40% 98%;
    --border: 214.3 31.8% 91.4%;
    --input: 214.3 31.8% 91.4%;
    --ring: 222.2 84% 4.9%;
    --radius: 0.5rem;
  }
  
  .dark {
    --background: 222.2 84% 4.9%;
    --foreground: 210 40% 98%;
    /* ... dark mode tokens */
  }
}
```

## Testing

### Unit Tests

Test components with React Testing Library:

```typescript
// components/__tests__/workspace-card.test.tsx
import { render, screen } from '@testing-library/react';
import { WorkspaceCard } from '../workspace-card';

describe('WorkspaceCard', () => {
  it('renders workspace information', () => {
    const workspace = {
      id: '1',
      name: 'Test Workspace',
      plan: 'pro',
    };
    
    render(<WorkspaceCard workspace={workspace} />);
    
    expect(screen.getByText('Test Workspace')).toBeInTheDocument();
    expect(screen.getByText('Pro Plan')).toBeInTheDocument();
  });
});
```

### Integration Tests

Test user flows with Playwright:

```typescript
// tests/workspaces.spec.ts
import { test, expect } from '@playwright/test';

test('create workspace flow', async ({ page }) => {
  await page.goto('/workspaces');
  
  // Click create button
  await page.click('button:has-text("Create Workspace")');
  
  // Fill form
  await page.fill('input[name="name"]', 'My New Workspace');
  await page.selectOption('select[name="plan"]', 'pro');
  
  // Submit
  await page.click('button[type="submit"]');
  
  // Verify creation
  await expect(page.locator('text=My New Workspace')).toBeVisible();
});
```

## Performance

### Code Splitting

Use dynamic imports for large components:

```typescript
// Lazy load heavy components
const MonacoEditor = dynamic(() => import('@/components/monaco-editor'), {
  loading: () => <div>Loading editor...</div>,
  ssr: false,
});
```

### Image Optimization

Use Next.js Image component:

```typescript
import Image from 'next/image';

export const Logo = () => (
  <Image
    src="/logo.png"
    alt="Hexabase AI"
    width={200}
    height={50}
    priority
  />
);
```

### React Query Optimization

Configure caching and refetching:

```typescript
// lib/react-query.ts
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000, // 5 minutes
      cacheTime: 10 * 60 * 1000, // 10 minutes
      refetchOnWindowFocus: false,
      retry: 3,
    },
  },
});
```

## Real-time Features

### WebSocket Integration

```typescript
// hooks/use-websocket.ts
export const useWebSocket = (url: string) => {
  const [socket, setSocket] = useState<WebSocket | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  
  useEffect(() => {
    const ws = new WebSocket(url);
    
    ws.onopen = () => {
      setIsConnected(true);
      setSocket(ws);
    };
    
    ws.onclose = () => {
      setIsConnected(false);
      setSocket(null);
    };
    
    return () => {
      ws.close();
    };
  }, [url]);
  
  return { socket, isConnected };
};
```

### Server-Sent Events

```typescript
// hooks/use-sse.ts
export const useSSE = (url: string) => {
  const [data, setData] = useState(null);
  
  useEffect(() => {
    const eventSource = new EventSource(url);
    
    eventSource.onmessage = (event) => {
      setData(JSON.parse(event.data));
    };
    
    return () => {
      eventSource.close();
    };
  }, [url]);
  
  return data;
};
```

## Error Handling

### Error Boundaries

```typescript
// components/error-boundary.tsx
export class ErrorBoundary extends React.Component<
  { children: React.ReactNode; fallback?: React.ComponentType<{ error: Error }> },
  { hasError: boolean; error?: Error }
> {
  constructor(props) {
    super(props);
    this.state = { hasError: false };
  }
  
  static getDerivedStateFromError(error: Error) {
    return { hasError: true, error };
  }
  
  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('Error caught by boundary:', error, errorInfo);
  }
  
  render() {
    if (this.state.hasError) {
      const Fallback = this.props.fallback || DefaultErrorFallback;
      return <Fallback error={this.state.error!} />;
    }
    
    return this.props.children;
  }
}
```

### API Error Handling

```typescript
// lib/api-client.ts
axios.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Handle unauthorized
      useAuthStore.getState().logout();
      router.push('/login');
    }
    
    // Show error toast
    toast.error(error.response?.data?.message || 'An error occurred');
    
    return Promise.reject(error);
  }
);
```

## Accessibility

### ARIA Labels

```typescript
export const IconButton = ({ icon, label, ...props }) => (
  <button aria-label={label} {...props}>
    {icon}
  </button>
);
```

### Keyboard Navigation

```typescript
export const Menu = ({ items }) => {
  const [selectedIndex, setSelectedIndex] = useState(0);
  
  const handleKeyDown = (e: React.KeyboardEvent) => {
    switch (e.key) {
      case 'ArrowDown':
        setSelectedIndex((i) => Math.min(i + 1, items.length - 1));
        break;
      case 'ArrowUp':
        setSelectedIndex((i) => Math.max(i - 1, 0));
        break;
      case 'Enter':
        items[selectedIndex].onClick();
        break;
    }
  };
  
  return (
    <div role="menu" onKeyDown={handleKeyDown}>
      {items.map((item, index) => (
        <div
          key={item.id}
          role="menuitem"
          tabIndex={0}
          aria-selected={index === selectedIndex}
        >
          {item.label}
        </div>
      ))}
    </div>
  );
};
```

## Best Practices

1. **Component Composition** - Build small, reusable components
2. **Type Safety** - Use TypeScript strictly, avoid `any`
3. **Performance** - Memoize expensive computations, use React.memo wisely
4. **Accessibility** - Test with screen readers, ensure keyboard navigation
5. **Error Handling** - Use error boundaries, handle loading states
6. **Testing** - Write tests for critical paths
7. **Code Organization** - Keep components focused and single-purpose
8. **State Management** - Use React Query for server state, Zustand for client state
9. **Styling** - Use Tailwind utilities, avoid inline styles
10. **Documentation** - Document complex components and hooks

## Debugging

### React DevTools

- Install React Developer Tools extension
- Use Profiler to identify performance issues
- Inspect component props and state

### Network Debugging

```typescript
// Enable request logging in development
if (process.env.NODE_ENV === 'development') {
  axios.interceptors.request.use((config) => {
    console.log('Request:', config.method?.toUpperCase(), config.url);
    return config;
  });
}
```

### Console Utilities

```typescript
// lib/debug.ts
export const debug = {
  log: (...args: any[]) => {
    if (process.env.NODE_ENV === 'development') {
      console.log('[DEBUG]', ...args);
    }
  },
  
  table: (data: any) => {
    if (process.env.NODE_ENV === 'development') {
      console.table(data);
    }
  },
};
```

## Resources

- [Next.js Documentation](https://nextjs.org/docs)
- [React Documentation](https://react.dev)
- [TypeScript Handbook](https://www.typescriptlang.org/docs/)
- [Tailwind CSS Documentation](https://tailwindcss.com/docs)
- [React Query Documentation](https://tanstack.com/query)
- [Zustand Documentation](https://github.com/pmndrs/zustand)# Development Guides

This directory contains guides for developers working on the Hexabase AI platform.

## 📚 Available Guides

### [Development Environment Setup](./dev-environment-setup.md)
Step-by-step guide to set up your local development environment, including:
- Prerequisites and required tools
- Quick setup with automated scripts
- Manual setup instructions
- IDE configuration
- Troubleshooting common issues

### [API Development Guide](./api-development-guide.md)
Best practices for developing the Go-based API, covering:
- Project structure and patterns
- Domain-driven design principles
- Error handling and validation
- Testing strategies
- Performance optimization

### [Frontend Development Guide](./frontend-development-guide.md)
Guide for Next.js frontend development, including:
- Component architecture
- State management with Zustand
- Data fetching with React Query
- UI components with shadcn/ui
- Testing with React Testing Library

### [Testing Guide](./testing-guide.md)
Comprehensive testing strategies:
- Unit testing patterns
- Integration testing
- End-to-end testing with Playwright
- Performance testing
- Test coverage requirements

### [Code Style Guide](./code-style-guide.md)
*Coming soon*

### [Git Workflow](./git-workflow.md)
*Coming soon*

## 🚀 Getting Started

If you're new to the project:

1. Start with [Development Environment Setup](./dev-environment-setup.md)
2. Review the guide specific to your area:
   - Backend: [API Development Guide](./api-development-guide.md)
   - Frontend: [Frontend Development Guide](./frontend-development-guide.md)
3. Understand our [Testing Guide](./testing-guide.md)
4. Check the main [Architecture Documentation](../../architecture/)

## 🛠️ Development Workflow

1. **Setup**: Follow the environment setup guide
2. **Branch**: Create a feature branch from `develop`
3. **Code**: Follow the relevant development guide
4. **Test**: Write and run tests as per testing guide
5. **Review**: Submit PR for code review
6. **Deploy**: Merge to `develop` triggers CI/CD

## 📝 Contributing

When contributing to these guides:

- Keep examples up-to-date with the codebase
- Test all code examples
- Include troubleshooting sections
- Link to relevant external documentation
- Update the README when adding new guides

## 🔗 Related Documentation

- [Architecture Overview](../../architecture/system-architecture.md)
- [API Reference](../../api-reference/README.md)
- [Deployment Guides](../deployment/)
- [Getting Started](../getting-started/)# API Development Guide

This guide covers best practices and patterns for developing the Hexabase AI API.

## Overview

The Hexabase AI API is built with:
- Go 1.24+
- Gin Web Framework
- GORM for database operations
- Wire for dependency injection
- OpenAPI/Swagger for documentation

## Project Structure

```
api/
├── cmd/
│   ├── api/          # API server entry point
│   └── worker/       # Background worker entry point
├── internal/
│   ├── api/          # HTTP handlers and routes
│   ├── domain/       # Business logic and models
│   ├── repository/   # Data access layer
│   ├── service/      # Service layer
│   └── infrastructure/ # Cross-cutting concerns
└── pkg/              # Public packages
```

## Development Patterns

### 1. Domain-Driven Design

We follow DDD principles with clear separation:
- **Domain Models**: Core business entities
- **Repositories**: Data access abstractions
- **Services**: Business logic implementation
- **Handlers**: HTTP request/response handling

### 2. Dependency Injection

Use Wire for compile-time dependency injection:

```go
// wire.go
var ServiceSet = wire.NewSet(
    NewUserService,
    wire.Bind(new(domain.UserService), new(*UserService)),
)
```

### 3. Error Handling

Consistent error handling across the API:

```go
type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details any    `json:"details,omitempty"`
}

func HandleError(c *gin.Context, err error) {
    var apiErr *APIError
    if errors.As(err, &apiErr) {
        c.JSON(apiErr.StatusCode(), apiErr)
        return
    }
    // Handle unknown errors
    c.JSON(500, APIError{Code: "INTERNAL_ERROR", Message: "Internal server error"})
}
```

### 4. Request Validation

Use binding tags for request validation:

```go
type CreateWorkspaceRequest struct {
    Name        string `json:"name" binding:"required,min=3,max=50"`
    Description string `json:"description" binding:"max=200"`
    Plan        string `json:"plan" binding:"required,oneof=starter pro enterprise"`
}
```

## API Conventions

### RESTful Endpoints

Follow REST conventions:
- `GET /api/v1/workspaces` - List resources
- `GET /api/v1/workspaces/:id` - Get single resource
- `POST /api/v1/workspaces` - Create resource
- `PUT /api/v1/workspaces/:id` - Update resource
- `DELETE /api/v1/workspaces/:id` - Delete resource

### Response Format

Consistent response structure:

```json
{
    "data": {
        "id": "123",
        "name": "My Workspace"
    },
    "meta": {
        "page": 1,
        "limit": 20,
        "total": 100
    }
}
```

### Status Codes

- `200 OK` - Successful GET, PUT
- `201 Created` - Successful POST
- `204 No Content` - Successful DELETE
- `400 Bad Request` - Validation errors
- `401 Unauthorized` - Authentication required
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource conflict
- `500 Internal Server Error` - Server errors

## Testing

### Unit Tests

Write tests for all services:

```go
func TestUserService_Create(t *testing.T) {
    // Setup
    mockRepo := mocks.NewMockUserRepository(t)
    service := NewUserService(mockRepo)
    
    // Test
    user := &domain.User{Name: "Test User"}
    mockRepo.On("Create", mock.Anything, user).Return(nil)
    
    err := service.Create(context.Background(), user)
    
    // Assert
    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
}
```

### Integration Tests

Test API endpoints:

```go
func TestCreateWorkspace(t *testing.T) {
    router := setupTestRouter()
    
    body := `{"name": "Test Workspace", "plan": "starter"}`
    req := httptest.NewRequest("POST", "/api/v1/workspaces", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    assert.Equal(t, 201, w.Code)
}
```

## Database Operations

### Migrations

Create migrations for schema changes:

```bash
migrate create -ext sql -dir migrations -seq create_users_table
```

### Transactions

Use transactions for data consistency:

```go
func (s *WorkspaceService) CreateWithProject(ctx context.Context, workspace *Workspace, project *Project) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        if err := tx.Create(workspace).Error; err != nil {
            return err
        }
        project.WorkspaceID = workspace.ID
        return tx.Create(project).Error
    })
}
```

## Security

### Authentication

JWT-based authentication:

```go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
            return
        }
        
        claims, err := ValidateToken(token)
        if err != nil {
            c.AbortWithStatusJSON(401, gin.H{"error": "Invalid token"})
            return
        }
        
        c.Set("user_id", claims.UserID)
        c.Next()
    }
}
```

### Authorization

Role-based access control:

```go
func RequireRole(roles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userRole := c.GetString("user_role")
        for _, role := range roles {
            if userRole == role {
                c.Next()
                return
            }
        }
        c.AbortWithStatusJSON(403, gin.H{"error": "Forbidden"})
    }
}
```

## Performance

### Database Queries

- Use eager loading to avoid N+1 queries
- Add indexes for frequently queried fields
- Use pagination for list endpoints
- Cache frequently accessed data

### Concurrent Operations

```go
func (s *Service) ProcessBatch(items []Item) error {
    errCh := make(chan error, len(items))
    var wg sync.WaitGroup
    
    for _, item := range items {
        wg.Add(1)
        go func(item Item) {
            defer wg.Done()
            if err := s.ProcessItem(item); err != nil {
                errCh <- err
            }
        }(item)
    }
    
    wg.Wait()
    close(errCh)
    
    // Check for errors
    for err := range errCh {
        if err != nil {
            return err
        }
    }
    return nil
}
```

## Monitoring

### Structured Logging

```go
logger.Info("workspace created",
    zap.String("workspace_id", workspace.ID),
    zap.String("user_id", userID),
    zap.String("plan", workspace.Plan),
)
```

### Metrics

```go
var (
    apiRequests = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "api_requests_total",
            Help: "Total number of API requests",
        },
        []string{"method", "endpoint", "status"},
    )
)
```

## OpenAPI Documentation

Document all endpoints:

```go
// @Summary Create workspace
// @Description Create a new workspace
// @Tags workspaces
// @Accept json
// @Produce json
// @Param workspace body CreateWorkspaceRequest true "Workspace data"
// @Success 201 {object} WorkspaceResponse
// @Failure 400 {object} APIError
// @Router /api/v1/workspaces [post]
func CreateWorkspace(c *gin.Context) {
    // Implementation
}
```

## Best Practices

1. **Keep handlers thin** - Move business logic to services
2. **Use contexts** - Pass context through all layers
3. **Handle errors gracefully** - Return meaningful error messages
4. **Validate input** - Use struct tags and custom validators
5. **Log appropriately** - Info for normal operations, Error for failures
6. **Write tests** - Aim for >80% coverage
7. **Document APIs** - Keep OpenAPI specs up to date
8. **Use transactions** - Ensure data consistency
9. **Implement idempotency** - For critical operations
10. **Monitor performance** - Add metrics and traces

## Common Patterns

### Repository Pattern

```go
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    GetByID(ctx context.Context, id string) (*User, error)
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id string) error
}
```

### Service Pattern

```go
type UserService struct {
    repo UserRepository
    logger *zap.Logger
}

func (s *UserService) Create(ctx context.Context, req *CreateUserRequest) (*User, error) {
    // Validation
    if err := req.Validate(); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    // Business logic
    user := &User{
        Name: req.Name,
        Email: req.Email,
    }
    
    // Persistence
    if err := s.repo.Create(ctx, user); err != nil {
        s.logger.Error("failed to create user", zap.Error(err))
        return nil, fmt.Errorf("failed to create user: %w", err)
    }
    
    return user, nil
}
```

## Debugging Tips

1. Use `gin.DebugMode` for development
2. Enable GORM debug mode for SQL queries
3. Use pprof for performance profiling
4. Add request IDs for tracing
5. Use structured logging for easier debugging

## Resources

- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Gin Documentation](https://gin-gonic.com/docs/)
- [GORM Documentation](https://gorm.io/docs/)
- [Wire Documentation](https://github.com/google/wire)# Testing Guide

Comprehensive testing guide for Hexabase AI platform covering unit tests, integration tests, and end-to-end tests.

## Overview

Our testing strategy includes:
- **Unit Tests**: Test individual components/functions
- **Integration Tests**: Test component interactions
- **End-to-End Tests**: Test complete user workflows
- **Performance Tests**: Ensure system meets performance requirements
- **Security Tests**: Validate security controls

Target: **80% code coverage** across all components

## Testing Stack

### Backend (Go)
- Testing framework: Built-in `testing` package
- Mocking: `testify/mock`
- Assertions: `testify/assert`
- HTTP testing: `httptest`
- Database testing: `sqlmock`

### Frontend (TypeScript/React)
- Test runner: Jest
- Component testing: React Testing Library
- E2E testing: Playwright
- Mocking: MSW (Mock Service Worker)

## Backend Testing

### Unit Tests

#### Service Layer Testing

```go
// internal/service/workspace/service_test.go
package workspace

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/hexabase/hexabase-ai/internal/domain/workspace"
    "github.com/hexabase/hexabase-ai/internal/domain/workspace/mocks"
)

func TestWorkspaceService_Create(t *testing.T) {
    // Arrange
    mockRepo := mocks.NewMockWorkspaceRepository(t)
    mockK8s := mocks.NewMockKubernetesRepository(t)
    service := NewService(mockRepo, mockK8s, nil)
    
    ctx := context.Background()
    ws := &workspace.Workspace{
        Name: "Test Workspace",
        Plan: workspace.PlanStarter,
    }
    
    // Set expectations
    mockRepo.On("Create", ctx, mock.MatchedBy(func(w *workspace.Workspace) bool {
        return w.Name == "Test Workspace"
    })).Return(nil).Once()
    
    mockK8s.On("CreateVCluster", ctx, mock.AnythingOfType("*workspace.Workspace")).
        Return(nil).Once()
    
    // Act
    err := service.Create(ctx, ws)
    
    // Assert
    assert.NoError(t, err)
    assert.NotEmpty(t, ws.ID)
    assert.Equal(t, workspace.StatusProvisioning, ws.Status)
    
    mockRepo.AssertExpectations(t)
    mockK8s.AssertExpectations(t)
}

func TestWorkspaceService_Create_ValidationError(t *testing.T) {
    // Test validation failures
    testCases := []struct {
        name      string
        workspace *workspace.Workspace
        wantError string
    }{
        {
            name:      "empty name",
            workspace: &workspace.Workspace{Name: "", Plan: workspace.PlanStarter},
            wantError: "workspace name is required",
        },
        {
            name:      "invalid plan",
            workspace: &workspace.Workspace{Name: "Test", Plan: "invalid"},
            wantError: "invalid workspace plan",
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            service := NewService(nil, nil, nil)
            err := service.Create(context.Background(), tc.workspace)
            assert.Error(t, err)
            assert.Contains(t, err.Error(), tc.wantError)
        })
    }
}
```

#### Repository Layer Testing

```go
// internal/repository/workspace/postgres_test.go
package workspace

import (
    "context"
    "testing"
    "time"
    
    "github.com/DATA-DOG/go-sqlmock"
    "github.com/stretchr/testify/assert"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func TestPostgresRepository_GetByID(t *testing.T) {
    // Setup mock database
    db, mock, err := sqlmock.New()
    assert.NoError(t, err)
    defer db.Close()
    
    gormDB, err := gorm.Open(postgres.New(postgres.Config{
        Conn: db,
    }), &gorm.Config{})
    assert.NoError(t, err)
    
    repo := NewPostgresRepository(gormDB)
    
    // Expected query
    workspaceID := "ws-123"
    rows := sqlmock.NewRows([]string{"id", "name", "plan", "status", "created_at"}).
        AddRow(workspaceID, "Test Workspace", "starter", "active", time.Now())
    
    mock.ExpectQuery("SELECT (.+) FROM \"workspaces\" WHERE id = ?").
        WithArgs(workspaceID).
        WillReturnRows(rows)
    
    // Execute
    workspace, err := repo.GetByID(context.Background(), workspaceID)
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, workspaceID, workspace.ID)
    assert.Equal(t, "Test Workspace", workspace.Name)
    assert.NoError(t, mock.ExpectationsWereMet())
}
```

#### HTTP Handler Testing

```go
// internal/api/handlers/workspaces_test.go
package handlers

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestWorkspaceHandler_Create(t *testing.T) {
    // Setup
    gin.SetMode(gin.TestMode)
    mockService := mocks.NewMockWorkspaceService(t)
    handler := NewWorkspaceHandler(mockService)
    
    // Create request
    reqBody := map[string]interface{}{
        "name": "Test Workspace",
        "plan": "starter",
    }
    body, _ := json.Marshal(reqBody)
    req := httptest.NewRequest("POST", "/api/v1/workspaces", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    
    // Mock expectations
    mockService.On("Create", mock.Anything, mock.MatchedBy(func(ws *workspace.Workspace) bool {
        return ws.Name == "Test Workspace" && ws.Plan == "starter"
    })).Return(nil).Once()
    
    // Execute
    w := httptest.NewRecorder()
    router := gin.New()
    router.POST("/api/v1/workspaces", handler.Create)
    router.ServeHTTP(w, req)
    
    // Assert
    assert.Equal(t, http.StatusCreated, w.Code)
    
    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.Equal(t, "Test Workspace", response["data"].(map[string]interface{})["name"])
    
    mockService.AssertExpectations(t)
}
```

### Integration Tests

```go
// tests/integration/workspace_flow_test.go
package integration

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/suite"
    "github.com/hexabase/hexabase-ai/internal/testutil"
)

type WorkspaceIntegrationSuite struct {
    suite.Suite
    testDB   *testutil.TestDatabase
    testK8s  *testutil.TestKubernetes
    app      *Application
}

func (s *WorkspaceIntegrationSuite) SetupSuite() {
    // Setup test database
    s.testDB = testutil.NewTestDatabase()
    s.testDB.Migrate()
    
    // Setup test Kubernetes
    s.testK8s = testutil.NewTestKubernetes()
    
    // Initialize application
    s.app = NewTestApplication(s.testDB.DB, s.testK8s.Client)
}

func (s *WorkspaceIntegrationSuite) TearDownSuite() {
    s.testDB.Cleanup()
    s.testK8s.Cleanup()
}

func (s *WorkspaceIntegrationSuite) TestCompleteWorkspaceLifecycle() {
    ctx := context.Background()
    
    // Create workspace
    ws := &workspace.Workspace{
        Name: "Integration Test",
        Plan: workspace.PlanPro,
    }
    err := s.app.WorkspaceService.Create(ctx, ws)
    s.NoError(err)
    s.NotEmpty(ws.ID)
    
    // Verify vCluster created
    vcluster, err := s.testK8s.GetVCluster(ws.ID)
    s.NoError(err)
    s.Equal("provisioning", vcluster.Status)
    
    // Simulate vCluster ready
    s.testK8s.UpdateVClusterStatus(ws.ID, "ready")
    
    // Create project in workspace
    project := &project.Project{
        WorkspaceID: ws.ID,
        Name:        "Test Project",
    }
    err = s.app.ProjectService.Create(ctx, project)
    s.NoError(err)
    
    // Verify namespace created in vCluster
    ns, err := s.testK8s.GetNamespace(ws.ID, project.ID)
    s.NoError(err)
    s.Equal(project.Name, ns.Labels["project-name"])
    
    // Delete workspace
    err = s.app.WorkspaceService.Delete(ctx, ws.ID)
    s.NoError(err)
    
    // Verify cleanup
    _, err = s.testK8s.GetVCluster(ws.ID)
    s.Error(err) // Should not exist
}

func TestWorkspaceIntegrationSuite(t *testing.T) {
    suite.Run(t, new(WorkspaceIntegrationSuite))
}
```

### Database Testing

```go
// internal/testutil/database.go
package testutil

import (
    "fmt"
    "testing"
    
    "github.com/ory/dockertest/v3"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

type TestDatabase struct {
    DB       *gorm.DB
    pool     *dockertest.Pool
    resource *dockertest.Resource
}

func NewTestDatabase() *TestDatabase {
    pool, err := dockertest.NewPool("")
    if err != nil {
        panic(err)
    }
    
    resource, err := pool.Run("postgres", "14", []string{
        "POSTGRES_PASSWORD=test",
        "POSTGRES_DB=test",
    })
    if err != nil {
        panic(err)
    }
    
    dsn := fmt.Sprintf("host=localhost port=%s user=postgres password=test dbname=test sslmode=disable",
        resource.GetPort("5432/tcp"))
    
    var db *gorm.DB
    if err := pool.Retry(func() error {
        var err error
        db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
        return err
    }); err != nil {
        panic(err)
    }
    
    return &TestDatabase{
        DB:       db,
        pool:     pool,
        resource: resource,
    }
}

func (td *TestDatabase) Cleanup() {
    td.pool.Purge(td.resource)
}
```

## Frontend Testing

### Component Tests

```typescript
// components/__tests__/workspace-list.test.tsx
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { WorkspaceList } from '../workspace-list';
import { server } from '@/mocks/server';
import { rest } from 'msw';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: false },
  },
});

const wrapper = ({ children }) => (
  <QueryClientProvider client={queryClient}>
    {children}
  </QueryClientProvider>
);

describe('WorkspaceList', () => {
  it('displays workspaces', async () => {
    render(<WorkspaceList />, { wrapper });
    
    // Wait for loading to complete
    await waitFor(() => {
      expect(screen.queryByText('Loading...')).not.toBeInTheDocument();
    });
    
    // Check workspaces are displayed
    expect(screen.getByText('Production Workspace')).toBeInTheDocument();
    expect(screen.getByText('Development Workspace')).toBeInTheDocument();
  });
  
  it('handles empty state', async () => {
    server.use(
      rest.get('/api/v1/workspaces', (req, res, ctx) => {
        return res(ctx.json({ data: [] }));
      })
    );
    
    render(<WorkspaceList />, { wrapper });
    
    await waitFor(() => {
      expect(screen.getByText('No workspaces found')).toBeInTheDocument();
    });
  });
  
  it('handles errors gracefully', async () => {
    server.use(
      rest.get('/api/v1/workspaces', (req, res, ctx) => {
        return res(ctx.status(500), ctx.json({ error: 'Server error' }));
      })
    );
    
    render(<WorkspaceList />, { wrapper });
    
    await waitFor(() => {
      expect(screen.getByText('Failed to load workspaces')).toBeInTheDocument();
    });
  });
});
```

### Hook Tests

```typescript
// hooks/__tests__/use-workspace.test.ts
import { renderHook, waitFor } from '@testing-library/react';
import { useWorkspace } from '../use-workspace';
import { wrapper } from '@/test-utils';

describe('useWorkspace', () => {
  it('fetches workspace data', async () => {
    const { result } = renderHook(() => useWorkspace('ws-123'), { wrapper });
    
    // Initially loading
    expect(result.current.isLoading).toBe(true);
    expect(result.current.data).toBeUndefined();
    
    // Wait for data
    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });
    
    // Check data
    expect(result.current.data).toEqual({
      id: 'ws-123',
      name: 'Test Workspace',
      plan: 'pro',
    });
  });
  
  it('handles workspace not found', async () => {
    const { result } = renderHook(() => useWorkspace('invalid'), { wrapper });
    
    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });
    
    expect(result.current.error?.message).toBe('Workspace not found');
  });
});
```

### E2E Tests with Playwright

```typescript
// tests/e2e/workspace-management.spec.ts
import { test, expect } from '@playwright/test';
import { mockAPI } from './helpers/mock-api';

test.describe('Workspace Management', () => {
  test.beforeEach(async ({ page }) => {
    await mockAPI(page);
    await page.goto('/login');
    await page.fill('[name=email]', 'test@example.com');
    await page.fill('[name=password]', 'password');
    await page.click('button[type=submit]');
    await page.waitForURL('/dashboard');
  });
  
  test('create new workspace', async ({ page }) => {
    // Navigate to workspaces
    await page.click('nav >> text=Workspaces');
    await page.waitForURL('/workspaces');
    
    // Click create button
    await page.click('button:has-text("Create Workspace")');
    
    // Fill form
    await page.fill('[name=name]', 'E2E Test Workspace');
    await page.fill('[name=description]', 'Created by Playwright test');
    await page.selectOption('[name=plan]', 'pro');
    
    // Submit
    await page.click('button[type=submit]');
    
    // Verify creation
    await expect(page.locator('text=E2E Test Workspace')).toBeVisible();
    await expect(page.locator('text=Pro Plan')).toBeVisible();
    
    // Take screenshot for visual regression
    await page.screenshot({ path: 'tests/screenshots/workspace-created.png' });
  });
  
  test('delete workspace', async ({ page }) => {
    await page.goto('/workspaces');
    
    // Find workspace card
    const workspaceCard = page.locator('[data-testid=workspace-card]', {
      hasText: 'Test Workspace'
    });
    
    // Open menu
    await workspaceCard.locator('button[aria-label="Options"]').click();
    
    // Click delete
    await page.click('text=Delete Workspace');
    
    // Confirm in dialog
    await page.click('dialog >> button:has-text("Delete")');
    
    // Verify deletion
    await expect(workspaceCard).not.toBeVisible();
  });
});
```

## Test Data Management

### Test Fixtures

```go
// internal/testutil/fixtures/workspace.go
package fixtures

import (
    "github.com/hexabase/hexabase-ai/internal/domain/workspace"
)

func NewWorkspace(opts ...func(*workspace.Workspace)) *workspace.Workspace {
    ws := &workspace.Workspace{
        ID:     "ws-test-123",
        Name:   "Test Workspace",
        Plan:   workspace.PlanStarter,
        Status: workspace.StatusActive,
    }
    
    for _, opt := range opts {
        opt(ws)
    }
    
    return ws
}

func WithPlan(plan string) func(*workspace.Workspace) {
    return func(ws *workspace.Workspace) {
        ws.Plan = plan
    }
}

func WithStatus(status string) func(*workspace.Workspace) {
    return func(ws *workspace.Workspace) {
        ws.Status = status
    }
}
```

### Test Factories

```typescript
// test-utils/factories.ts
import { Factory } from 'fishery';
import { faker } from '@faker-js/faker';
import { Workspace, Project, User } from '@/types';

export const workspaceFactory = Factory.define<Workspace>(() => ({
  id: faker.string.uuid(),
  name: faker.company.name(),
  description: faker.company.catchPhrase(),
  plan: faker.helpers.arrayElement(['starter', 'pro', 'enterprise']),
  status: 'active',
  createdAt: faker.date.past().toISOString(),
  updatedAt: faker.date.recent().toISOString(),
}));

export const projectFactory = Factory.define<Project>(() => ({
  id: faker.string.uuid(),
  workspaceId: faker.string.uuid(),
  name: faker.commerce.productName(),
  description: faker.commerce.productDescription(),
  createdAt: faker.date.past().toISOString(),
  updatedAt: faker.date.recent().toISOString(),
}));

export const userFactory = Factory.define<User>(() => ({
  id: faker.string.uuid(),
  email: faker.internet.email(),
  name: faker.person.fullName(),
  role: faker.helpers.arrayElement(['admin', 'developer', 'viewer']),
  createdAt: faker.date.past().toISOString(),
}));
```

## Performance Testing

### Load Testing with k6

```javascript
// tests/performance/workspace-api.js
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

export const options = {
  stages: [
    { duration: '30s', target: 10 },  // Ramp up
    { duration: '1m', target: 50 },   // Stay at 50 users
    { duration: '30s', target: 100 }, // Spike to 100
    { duration: '1m', target: 100 },  // Stay at 100
    { duration: '30s', target: 0 },   // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests under 500ms
    errors: ['rate<0.1'],             // Error rate under 10%
  },
};

const BASE_URL = 'https://api.hexabase.ai';

export function setup() {
  // Login and get token
  const loginRes = http.post(`${BASE_URL}/auth/login`, {
    email: 'loadtest@example.com',
    password: 'loadtest123',
  });
  
  return { token: loginRes.json('token') };
}

export default function (data) {
  const headers = {
    'Authorization': `Bearer ${data.token}`,
    'Content-Type': 'application/json',
  };
  
  // List workspaces
  const listRes = http.get(`${BASE_URL}/api/v1/workspaces`, { headers });
  check(listRes, {
    'list status is 200': (r) => r.status === 200,
    'list duration < 500ms': (r) => r.timings.duration < 500,
  }) || errorRate.add(1);
  
  // Create workspace
  const createRes = http.post(
    `${BASE_URL}/api/v1/workspaces`,
    JSON.stringify({
      name: `Load Test ${Date.now()}`,
      plan: 'starter',
    }),
    { headers }
  );
  
  check(createRes, {
    'create status is 201': (r) => r.status === 201,
    'create duration < 1000ms': (r) => r.timings.duration < 1000,
  }) || errorRate.add(1);
  
  sleep(1);
}
```

## Security Testing

### Authentication Tests

```go
func TestAuthMiddleware_ValidToken(t *testing.T) {
    // Create valid JWT
    token, err := auth.GenerateToken("user-123", "admin")
    assert.NoError(t, err)
    
    // Create request with token
    req := httptest.NewRequest("GET", "/api/v1/workspaces", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    
    // Test middleware
    w := httptest.NewRecorder()
    router := gin.New()
    router.Use(AuthMiddleware())
    router.GET("/api/v1/workspaces", func(c *gin.Context) {
        userID, _ := c.Get("user_id")
        c.JSON(200, gin.H{"user_id": userID})
    })
    
    router.ServeHTTP(w, req)
    
    assert.Equal(t, 200, w.Code)
    assert.Contains(t, w.Body.String(), "user-123")
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
    testCases := []struct {
        name   string
        token  string
        status int
    }{
        {"missing token", "", 401},
        {"invalid format", "invalid", 401},
        {"expired token", generateExpiredToken(), 401},
        {"wrong signature", "Bearer wrong.token.here", 401},
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            req := httptest.NewRequest("GET", "/api/v1/workspaces", nil)
            if tc.token != "" {
                req.Header.Set("Authorization", tc.token)
            }
            
            w := httptest.NewRecorder()
            router := gin.New()
            router.Use(AuthMiddleware())
            router.GET("/api/v1/workspaces", func(c *gin.Context) {
                c.JSON(200, gin.H{})
            })
            
            router.ServeHTTP(w, req)
            assert.Equal(t, tc.status, w.Code)
        })
    }
}
```

## Test Coverage

### Running Coverage

```bash
# Backend coverage
cd api
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Frontend coverage
cd ui
npm run test:coverage
```

### Coverage Requirements

- Overall: 80% minimum
- Critical paths: 90% minimum
- New code: 85% minimum

## CI/CD Integration

### GitHub Actions

```yaml
# .github/workflows/test.yml
name: Tests

on: [push, pull_request]

jobs:
  backend-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24'
      
      - name: Run tests
        run: |
          cd api
          go test -v -race -coverprofile=coverage.out ./...
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./api/coverage.out
  
  frontend-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Node
        uses: actions/setup-node@v3
        with:
          node-version: '18'
      
      - name: Install dependencies
        run: |
          cd ui
          npm ci
      
      - name: Run tests
        run: |
          cd ui
          npm run test:ci
      
      - name: Run E2E tests
        run: |
          cd ui
          npx playwright install
          npm run test:e2e
```

## Best Practices

1. **Test Naming**: Use descriptive test names that explain what is being tested
2. **Test Independence**: Each test should be independent and not rely on others
3. **Mock External Dependencies**: Use mocks for databases, APIs, and external services
4. **Test Data**: Use factories and fixtures for consistent test data
5. **Assertions**: Make specific assertions, avoid generic checks
6. **Coverage**: Aim for high coverage but focus on critical paths
7. **Performance**: Keep tests fast, parallelize where possible
8. **Maintenance**: Refactor tests alongside code changes
9. **Documentation**: Document complex test scenarios
10. **CI/CD**: Run tests automatically on every commit

## Troubleshooting

### Common Issues

1. **Flaky Tests**
   - Add proper waits and retries
   - Ensure proper test isolation
   - Mock time-dependent operations

2. **Slow Tests**
   - Use test database containers
   - Parallelize test execution
   - Mock expensive operations

3. **Coverage Gaps**
   - Focus on untested critical paths
   - Add tests for error scenarios
   - Test edge cases

## Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Testify Documentation](https://github.com/stretchr/testify)
- [React Testing Library](https://testing-library.com/docs/react-testing-library/intro/)
- [Playwright Documentation](https://playwright.dev/)
- [k6 Documentation](https://k6.io/docs/)# Operations Guide

This section is designed for Infrastructure and DevOps teams responsible for deploying, maintaining, and operating Hexabase KaaS in production environments.

## In This Section

Most deployment and operations guides have been moved to the [Deployment](../deployment/) section for better organization.

### Available Here

This directory contains additional operational procedures and policies.

### Deployment Guides

For deployment-related documentation, please see:

- [Kubernetes Deployment](../deployment/kubernetes-deployment.md) - Basic Kubernetes deployment
- [Production Setup](../deployment/production-setup.md) - Production-grade deployment guide
- [Monitoring Setup](../deployment/monitoring-setup.md) - Monitoring and observability
- [Backup & Recovery](../deployment/backup-recovery.md) - Backup and disaster recovery
- [Deployment Policies](../deployment/deployment-policies.md) - Organizational policies

### Troubleshooting Guide
*Coming soon* - Common operational issues and their solutions.

## Quick Start for Operations

1. **Review prerequisites**: Check system requirements below
2. **Prepare infrastructure**: Follow [Production Setup](../deployment/production-setup.md)
3. **Deploy the platform**: Use [Kubernetes Deployment](../deployment/kubernetes-deployment.md)
4. **Set up monitoring**: Configure [Monitoring Setup](../deployment/monitoring-setup.md)
5. **Plan for disasters**: Implement [Backup & Recovery](../deployment/backup-recovery.md)

## Infrastructure Requirements

### Minimum Production Requirements

- **Kubernetes Cluster**: v1.24+ (or K3s v1.24+)
- **Nodes**: 3+ control plane, 3+ worker nodes
- **Resources per node**:
  - CPU: 4+ cores
  - Memory: 16GB+
  - Storage: 100GB+ SSD
- **Network**: 10Gbps interconnect recommended

### External Dependencies

- **PostgreSQL**: v14+ (RDS/Cloud SQL recommended)
- **Redis**: v6+ (ElastiCache/Cloud Memorystore)
- **Object Storage**: S3-compatible for backups
- **Load Balancer**: For API and ingress
- **DNS**: For service discovery

## Security Considerations

- Network policies for tenant isolation
- RBAC configuration
- Secret management (Vault/Sealed Secrets)
- TLS everywhere
- Regular security updates

## Compliance & Governance

- Audit logging
- Data residency requirements
- Backup retention policies
- Access control procedures

## Getting Help

- Review logs and metrics first
- Check troubleshooting sections in deployment guides
- Contact support with diagnostic bundle
- Visit [Community Discord](https://discord.gg/hexabase)# Hexabase AI Overview

Welcome to Hexabase AI - the next-generation Kubernetes-as-a-Service platform with built-in AI capabilities.

## What is Hexabase AI?

Hexabase AI is a multi-tenant Kubernetes platform that simplifies application deployment and management while providing powerful AI-driven automation features. Built on K3s and vCluster technology, it offers isolated Kubernetes environments with enterprise-grade security and scalability.

## Key Features

### 🚀 Instant Kubernetes Environments
- **Workspaces**: Isolated vCluster environments provisioned in seconds
- **Projects**: Namespace-based resource organization
- **Auto-scaling**: Intelligent resource management

### 🤖 AI-Powered Operations
- **Smart Troubleshooting**: AI agents analyze and fix issues
- **Code Generation**: Generate Kubernetes manifests and configurations
- **Performance Optimization**: AI-driven resource recommendations

### 🔧 Developer-Friendly
- **Simple CLI**: Intuitive command-line interface
- **Web Dashboard**: Modern UI for visual management
- **API-First**: Complete REST and WebSocket APIs

### 🏗️ Enterprise Ready
- **Multi-tenancy**: Complete isolation between workspaces
- **RBAC**: Fine-grained access control
- **Compliance**: SOC2, HIPAA, GDPR ready
- **High Availability**: Built-in redundancy and failover

### 💼 Built-in Services
- **CronJobs**: Scheduled task management
- **Serverless Functions**: Event-driven compute with Knative
- **Backup & Restore**: Automated data protection
- **Monitoring**: Integrated Prometheus and Grafana

## Use Cases

### Development Teams
- Spin up isolated development environments
- Test applications in production-like settings
- Collaborate with built-in access controls

### DevOps Engineers
- Automate deployment pipelines
- Manage multiple environments from one place
- Monitor and optimize resource usage

### Enterprises
- Provide self-service Kubernetes to teams
- Maintain compliance and security standards
- Reduce infrastructure costs

### AI/ML Engineers
- Deploy ML models as serverless functions
- Schedule training jobs with CronJobs
- Use AI agents for automated operations

## How It Works

1. **Create Organization**: Set up your billing and team unit
2. **Provision Workspace**: Get an isolated Kubernetes environment
3. **Deploy Applications**: Use kubectl, UI, or API
4. **Monitor & Scale**: Built-in observability and auto-scaling
5. **Collaborate**: Invite team members with role-based access

## Platform Architecture

```
┌─────────────────────┐
│   User Interface    │
│  (Web UI / CLI)     │
└──────────┬──────────┘
           │
┌──────────▼──────────┐
│   Hexabase API      │
│  (Control Plane)    │
└──────────┬──────────┘
           │
┌──────────▼──────────┐     ┌─────────────────┐
│   Host K3s Cluster  │────▶│  vCluster       │
│                     │     │  (Workspace)    │
└─────────────────────┘     └─────────────────┘
           │
┌──────────▼──────────┐
│   Infrastructure    │
│ (Storage, Network)  │
└─────────────────────┘
```

## Pricing Plans

### Starter
- 1 workspace
- 2 CPU, 4GB RAM
- Community support
- Perfect for individuals

### Pro
- 5 workspaces
- 8 CPU, 32GB RAM
- Email support
- Great for small teams

### Enterprise
- Unlimited workspaces
- Custom resources
- 24/7 support
- SLA guarantees

## Getting Started

Ready to begin? Follow our [Quick Start Guide](./quick-start.md) to deploy your first application in minutes!

## Compare to Alternatives

| Feature | Hexabase AI | Traditional K8s | Other KaaS |
|---------|-------------|-----------------|------------|
| Setup Time | < 1 minute | Hours/Days | 10-30 minutes |
| Multi-tenancy | Built-in | Complex setup | Limited |
| AI Operations | ✅ | ❌ | ❌ |
| Cost | Pay-per-use | High fixed | Variable |
| Learning Curve | Low | High | Medium |

## Next Steps

1. [Understand Core Concepts](./concepts.md)
2. [Follow Quick Start Guide](./quick-start.md)
3. [Explore Features](../../architecture/system-architecture.md)
4. [Join Community](https://discord.gg/hexabase)

## Questions?

- **Sales**: sales@hexabase.ai
- **Support**: support@hexabase.ai
- **Documentation**: [docs.hexabase.ai](https://docs.hexabase.ai)
- **Status**: [status.hexabase.ai](https://status.hexabase.ai)

---

*Hexabase AI - Kubernetes Made Simple, Powered by AI* 🚀# Hexabase AI: Concept and Architecture

## 1. Project Overview

### Vision

Hexabase AI is an open-source, multi-tenant Kubernetes as a Service platform built on K3s and vCluster, designed to make Kubernetes accessible to developers of all skill levels.

### Core Values

- **Ease of Adoption**: Lightweight K3s base with vCluster virtualization for rapid deployment
- **Intuitive UX**: Abstract Kubernetes complexity through Organizations, Workspaces, and Projects
- **Strong Tenant Isolation**: vCluster provides dedicated API servers and control planes per tenant
- **Cloud-Native Operations**: Built-in Prometheus, Grafana, Loki monitoring; Flux GitOps; Kyverno policies
- **Open Source Transparency**: Community-driven development with full source code availability

### Existing Codebases

- **UI (Next.js)**: https://github.com/b-eee/hxb-next-webui
- **API (Go)**: https://github.com/b-eee/apicore

Both repositories require significant reimplementation based on this specification.

## 2. System Architecture

### Component Overview

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  Hexabase UI    │────▶│  Hexabase API    │────▶│  Host K3s       │
│  (Next.js)      │     │  (Control Plane) │     │  Cluster        │
└─────────────────┘     └────────┬─────────┘     └────────┬────────┘
                                 │                        │
                        ┌────────┴────────┐               │
                        │                 │          ┌────▼────────┐
                  ┌─────▼─────┐     ┌─────▼─────┐    │  vClusters  │
                  │PostgreSQL │     │   Redis   │    │  (Tenants)  │
                  └───────────┘     └───────────┘    └─────────────┘
                                          │
                                    ┌─────▼─────┐
                                    │   NATS    │
                                    └───────────┘
```

### Data Flow

1. **User Operations**: Browser → UI → API requests with auth tokens
2. **API Processing**: Auth validation → Business logic → DB updates → Async tasks
3. **vCluster Orchestration**: API → client-go → Host K3s → vCluster lifecycle
4. **Async Processing**: API → NATS → Workers → Long-running operations
5. **State Persistence**: PostgreSQL for all entities, Redis for caching
6. **Monitoring**: Prometheus metrics → Loki logs → Grafana dashboards
7. **GitOps Deployment**: Git → Flux → Host K3s → Automated updates
8. **Policy Enforcement**: Kyverno admission controller → Policy validation

## 3. Core Concepts

| Hexabase Concept      | Kubernetes Equivalent | Scope     | Description                               |
| --------------------- | --------------------- | --------- | ----------------------------------------- |
| Organization          | N/A                   | Hexabase  | Billing and user management unit          |
| Workspace             | vCluster              | Host K3s  | Isolated Kubernetes environment           |
| Workspace Plan        | ResourceQuota/Nodes   | vCluster  | Resource limits and node allocation       |
| Organization User     | N/A                   | Hexabase  | Billing/admin personnel                   |
| Workspace Member      | OIDC Subject          | vCluster  | Technical users with kubectl access       |
| Workspace Group       | OIDC Claim            | vCluster  | Permission assignment unit (hierarchical) |
| Workspace ClusterRole | ClusterRole           | vCluster  | Preset workspace-wide permissions         |
| Project               | Namespace             | vCluster  | Resource isolation within workspace       |
| Project Role          | Role                  | Namespace | Custom permissions within project         |

## 4. User Flows

### 4.1 Signup and Organization Management

- **Authentication**: External IdP (Google/GitHub) via OIDC
- **Auto-provisioning**: Private Organization created on first signup
- **Organization Admin**: Manages billing (Stripe) and invites users

### 4.2 Workspace (vCluster) Management

- **Creation**: Select plan → Provision vCluster → Configure OIDC
- **Initial Setup**:
  - Auto-create ClusterRoles: `hexabase:workspace-admin`, `hexabase:workspace-viewer`
  - Create default groups: `WorkspaceMembers` → `WSAdmins`, `WSUsers`
  - Assign creator to `WSAdmins` group

### 4.3 Project (Namespace) Management

- **Creation**: UI request → Create namespace in vCluster
- **ResourceQuota**: Auto-apply based on workspace plan
- **Custom Roles**: Create project-scoped roles via UI

### 4.4 Permission Management

- **Assignment**: Groups → Roles/ClusterRoles via UI
- **Inheritance**: Recursive group membership resolution
- **OIDC Integration**: Flattened groups in token claims

## 5. Technology Stack

### Core Components

- **Frontend**: Next.js
- **Backend**: Go (Golang)
- **Database**: PostgreSQL (primary), Redis (cache)
- **Messaging**: NATS
- **Container Platform**: K3s + vCluster

### CI/CD & Operations

- **CI Pipeline**: Tekton (Kubernetes-native)
- **GitOps**: ArgoCD or Flux
- **Container Scanning**: Trivy
- **Runtime Security**: Falco
- **Policy Engine**: Kyverno

## 6. Installation (IaC)

### Helm Umbrella Chart

```yaml
apiVersion: v2
name: hexabase-ai
dependencies:
  - name: postgresql
    repository: https://charts.bitnami.com/bitnami
  - name: redis
    repository: https://charts.bitnami.com/bitnami
  - name: nats
    repository: https://nats-io.github.io/k8s/helm/charts/
```

### Quick Install

```bash
helm repo add hexabase https://hexabase.ai/charts
helm install hexabase-ai hexabase/hexabase-ai -f values.yaml
```

## 7. Key Features

### Multi-tenancy

- vCluster provides complete API server isolation
- Dedicated control plane components per tenant
- Optional dedicated nodes for premium plans

### Security

- External IdP authentication only
- Hexabase acts as OIDC provider for vClusters
- Kyverno policy enforcement
- Network isolation between tenants

### Scalability

- Horizontal scaling of control plane components
- Queue-based async processing
- Stateless API design
- Redis caching layer

### Observability

- Prometheus metrics collection
- Centralized logging with Loki
- Pre-built Grafana dashboards
- Real-time resource usage tracking

## 8. Summary

Hexabase AI democratizes Kubernetes access through intelligent abstractions, strong multi-tenancy, and enterprise-grade operations tooling. By leveraging K3s and vCluster, it provides a production-ready platform that scales from individual developers to large organizations, all while maintaining the flexibility and power of native Kubernetes.

The open-source nature ensures transparency, community-driven innovation, and the ability to customize for specific requirements. With simple Helm-based installation and comprehensive monitoring, Hexabase AI represents a new standard for accessible Kubernetes platforms.
# Quick Start Guide

Get up and running with Hexabase AI in minutes.

## Prerequisites

Before you begin, ensure you have:
- A Hexabase AI account
- kubectl installed on your machine
- Basic knowledge of Kubernetes concepts

## 1. Sign Up and Login

### Create an Account

1. Visit [https://app.hexabase.ai](https://app.hexabase.ai)
2. Click "Sign Up"
3. Choose your authentication method:
   - Google OAuth
   - GitHub OAuth
   - Email/Password

### First Login

After signing up, you'll be guided through the initial setup:

1. **Create Organization**: Your billing and team management unit
2. **Choose Plan**: Select from Starter, Pro, or Enterprise
3. **Create First Workspace**: Your isolated Kubernetes environment

## 2. Create Your First Workspace

Workspaces are isolated Kubernetes environments (vClusters) for your applications.

### Using the UI

1. Navigate to the Workspaces page
2. Click "Create Workspace"
3. Fill in the details:
   ```
   Name: my-first-workspace
   Description: Getting started with Hexabase AI
   Plan: Starter (1 CPU, 2GB RAM)
   ```
4. Click "Create"

The workspace will be provisioned in about 30 seconds.

### Using the CLI

```bash
# Install Hexabase CLI
curl -sSL https://get.hexabase.ai | sh

# Login
hks login

# Create workspace
hks workspace create my-first-workspace --plan starter
```

## 3. Connect to Your Workspace

Once your workspace is ready, you can connect using kubectl.

### Download Kubeconfig

1. Go to your workspace dashboard
2. Click "Download Kubeconfig"
3. Save the file to `~/.kube/config-hexabase`

### Configure kubectl

```bash
# Set the kubeconfig
export KUBECONFIG=~/.kube/config-hexabase

# Verify connection
kubectl get nodes
```

## 4. Deploy Your First Application

Let's deploy a simple web application.

### Create a Project (Namespace)

Projects organize resources within a workspace.

```bash
# Using CLI
hks project create my-app --workspace my-first-workspace

# Or using kubectl
kubectl create namespace my-app
```

### Deploy an Application

Create a file named `app.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hello-world
  namespace: my-app
spec:
  replicas: 2
  selector:
    matchLabels:
      app: hello-world
  template:
    metadata:
      labels:
        app: hello-world
    spec:
      containers:
      - name: hello
        image: nginx:alpine
        ports:
        - containerPort: 80
        resources:
          requests:
            memory: "64Mi"
            cpu: "250m"
          limits:
            memory: "128Mi"
            cpu: "500m"
---
apiVersion: v1
kind: Service
metadata:
  name: hello-world
  namespace: my-app
spec:
  selector:
    app: hello-world
  ports:
  - port: 80
    targetPort: 80
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: hello-world
  namespace: my-app
spec:
  rules:
  - host: hello.my-first-workspace.hexabase.app
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: hello-world
            port:
              number: 80
```

Deploy the application:

```bash
kubectl apply -f app.yaml
```

### Access Your Application

Your application will be available at:
```
https://hello.my-first-workspace.hexabase.app
```

## 5. Manage Your Application

### View Application Status

```bash
# Check deployment status
kubectl get deployments -n my-app

# View pods
kubectl get pods -n my-app

# Check service
kubectl get svc -n my-app
```

### Scale Your Application

```bash
# Scale to 5 replicas
kubectl scale deployment hello-world -n my-app --replicas=5

# Or use the UI
# Navigate to Applications > hello-world > Scale
```

### View Logs

```bash
# View logs from all pods
kubectl logs -n my-app -l app=hello-world

# Stream logs
kubectl logs -n my-app -l app=hello-world -f
```

## 6. Set Up Monitoring

Hexabase AI includes built-in monitoring.

### Access Grafana Dashboard

1. Go to your workspace dashboard
2. Click "Monitoring"
3. View metrics for:
   - CPU usage
   - Memory usage
   - Network traffic
   - Request rates

### Set Up Alerts

1. Navigate to Monitoring > Alerts
2. Click "Create Alert"
3. Configure alert conditions:
   ```
   Metric: CPU Usage
   Condition: > 80%
   Duration: 5 minutes
   Action: Send email
   ```

## 7. Deploy a CronJob

Schedule recurring tasks using CronJobs.

### Create a Backup CronJob

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: backup-job
  namespace: my-app
spec:
  schedule: "0 2 * * *"  # Daily at 2 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: your-backup-image:latest
            command:
            - /bin/sh
            - -c
            - echo "Running backup..." && date
          restartPolicy: OnFailure
```

Deploy:

```bash
kubectl apply -f backup-cronjob.yaml
```

## 8. Use AI Assistant

Hexabase AI includes an AI assistant for troubleshooting.

### Ask for Help

```bash
# Using CLI
hks ai "Why is my deployment failing?"

# The AI will analyze your resources and provide suggestions
```

### Example AI Commands

- "Help me optimize my resource limits"
- "Why is my pod crashing?"
- "Generate a Dockerfile for a Node.js app"
- "Create a CI/CD pipeline for my application"

## 9. Manage Team Access

### Invite Team Members

1. Go to Organization Settings
2. Click "Team Members"
3. Click "Invite Member"
4. Enter email and select role:
   - **Admin**: Full access
   - **Developer**: Deploy and manage applications
   - **Viewer**: Read-only access

### Set Workspace Permissions

1. Navigate to Workspace > Settings > Access
2. Add team members
3. Assign roles specific to this workspace

## 10. Next Steps

### Explore Advanced Features

- **Serverless Functions**: Deploy event-driven functions
- **Backup & Restore**: Set up automated backups
- **CI/CD Integration**: Connect your Git repository
- **Custom Domains**: Use your own domain names
- **Multi-region Deployment**: Deploy across regions

### Learn More

- [Core Concepts](./concepts.md)
- [API Documentation](../../api-reference/README.md)
- [Video Tutorials](https://hexabase.ai/tutorials)
- [Community Forum](https://community.hexabase.ai)

### Get Help

- **Documentation**: [docs.hexabase.ai](https://docs.hexabase.ai)
- **Support**: support@hexabase.ai
- **Discord**: [Join our Discord](https://discord.gg/hexabase)
- **Status Page**: [status.hexabase.ai](https://status.hexabase.ai)

## Troubleshooting

### Common Issues

**Cannot connect to workspace**
```bash
# Verify kubeconfig
kubectl config current-context

# Test connection
kubectl cluster-info
```

**Application not accessible**
```bash
# Check ingress
kubectl get ingress -n my-app

# Verify DNS
nslookup hello.my-first-workspace.hexabase.app
```

**Out of resources**
```bash
# Check resource usage
kubectl top nodes
kubectl top pods -n my-app

# View quotas
kubectl get resourcequota -n my-app
```

---

🎉 **Congratulations!** You've successfully deployed your first application on Hexabase AI. Continue exploring our platform's features to build and scale your applications with ease.# Getting Started

This section provides an introduction to Hexabase KaaS and helps you understand the platform's core concepts.

## In This Section

### [Overview](./overview.md)
Get a high-level understanding of what Hexabase KaaS is and its key features.

### [Quick Start](./quick-start.md)
Follow our step-by-step guide to get Hexabase KaaS running in your environment.

### [Core Concepts](./concepts.md)
Learn about the fundamental concepts and terminology used throughout Hexabase KaaS.

## What is Hexabase KaaS?

Hexabase KaaS (Kubernetes as a Service) is a multi-tenant platform that provides isolated Kubernetes environments using vCluster technology. It enables organizations to:

- **Create isolated Kubernetes environments** for different teams and projects
- **Manage resources** with fine-grained quotas and limits
- **Integrate with existing identity providers** via OAuth2/OIDC
- **Monitor and observe** cluster health and resource usage
- **Handle billing** with flexible subscription plans

## Who Should Read This?

- **Developers** looking to understand the platform architecture
- **DevOps Engineers** preparing to deploy Hexabase KaaS
- **Platform Administrators** managing multi-tenant Kubernetes environments
- **Product Managers** evaluating Kubernetes-as-a-Service solutions

## Next Steps

1. Read the [Overview](./overview.md) to understand the platform
2. Follow the [Quick Start](./quick-start.md) guide
3. Deep dive into [Core Concepts](./concepts.md)
4. Set up your [Development Environment](../development/dev-environment-setup.md)# Production Setup Guide

This guide covers deploying Hexabase AI platform in a production environment.

## Overview

A production Hexabase AI deployment consists of:
- Host K3s cluster for the control plane
- PostgreSQL database cluster
- Redis cluster for caching
- NATS for messaging
- Object storage (S3-compatible)
- Load balancers and ingress controllers
- Monitoring and logging infrastructure

## Prerequisites

### Infrastructure Requirements

#### Minimum Production Setup (Small)
- **Control Plane**: 3 nodes (4 CPU, 16GB RAM each)
- **Worker Nodes**: 3 nodes (8 CPU, 32GB RAM each)
- **Database**: 3 nodes (4 CPU, 16GB RAM, 500GB SSD each)
- **Storage**: 1TB S3-compatible object storage

#### Recommended Production Setup (Medium)
- **Control Plane**: 3 nodes (8 CPU, 32GB RAM each)
- **Worker Nodes**: 5+ nodes (16 CPU, 64GB RAM each)
- **Database**: 3 nodes (8 CPU, 32GB RAM, 1TB NVMe each)
- **Cache**: 3 Redis nodes (4 CPU, 16GB RAM each)
- **Storage**: 10TB S3-compatible object storage

### Software Requirements
- Ubuntu 22.04 LTS or RHEL 8+
- K3s v1.28+
- PostgreSQL 15+
- Redis 7+
- NATS 2.10+

## Installation Steps

### 1. Prepare Infrastructure

#### Set Up Nodes

```bash
# On all nodes
sudo apt-get update
sudo apt-get install -y curl wget git

# Configure kernel parameters
cat <<EOF | sudo tee /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-iptables = 1
net.ipv4.ip_forward = 1
net.bridge.bridge-nf-call-ip6tables = 1
EOF

sudo sysctl --system

# Disable swap
sudo swapoff -a
sudo sed -i '/ swap / s/^/#/' /etc/fstab
```

#### Configure Firewall

```bash
# Control plane nodes
sudo ufw allow 6443/tcp  # Kubernetes API
sudo ufw allow 2379:2380/tcp  # etcd
sudo ufw allow 10250/tcp  # Kubelet API
sudo ufw allow 10251/tcp  # kube-scheduler
sudo ufw allow 10252/tcp  # kube-controller-manager

# Worker nodes
sudo ufw allow 10250/tcp  # Kubelet API
sudo ufw allow 30000:32767/tcp  # NodePort Services

# All nodes
sudo ufw allow 8472/udp  # Flannel VXLAN
sudo ufw allow 51820/udp  # WireGuard
```

### 2. Install K3s Cluster

#### Install First Control Plane Node

```bash
# On first control plane node
curl -sfL https://get.k3s.io | sh -s - server \
  --cluster-init \
  --disable traefik \
  --disable servicelb \
  --write-kubeconfig-mode 644 \
  --node-taint CriticalAddonsOnly=true:NoExecute \
  --etcd-expose-metrics true
  
# Get node token
sudo cat /var/lib/rancher/k3s/server/node-token
```

#### Join Additional Control Plane Nodes

```bash
# On other control plane nodes
curl -sfL https://get.k3s.io | sh -s - server \
  --server https://<first-node-ip>:6443 \
  --token <node-token> \
  --disable traefik \
  --disable servicelb \
  --node-taint CriticalAddonsOnly=true:NoExecute
```

#### Join Worker Nodes

```bash
# On worker nodes
curl -sfL https://get.k3s.io | sh -s - agent \
  --server https://<control-plane-ip>:6443 \
  --token <node-token>
```

### 3. Install Core Components

#### Install Helm

```bash
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
```

#### Install NGINX Ingress Controller

```bash
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update

helm install ingress-nginx ingress-nginx/ingress-nginx \
  --namespace ingress-nginx \
  --create-namespace \
  --set controller.service.type=LoadBalancer \
  --set controller.metrics.enabled=true \
  --set controller.podAnnotations."prometheus\.io/scrape"=true
```

#### Install Cert-Manager

```bash
helm repo add jetstack https://charts.jetstack.io
helm repo update

helm install cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --set installCRDs=true \
  --set prometheus.enabled=true
```

### 4. Deploy PostgreSQL Cluster

#### Using CloudNativePG

```bash
kubectl apply -f https://raw.githubusercontent.com/cloudnative-pg/cloudnative-pg/release-1.22/releases/cnpg-1.22.0.yaml

# Create PostgreSQL cluster
cat <<EOF | kubectl apply -f -
apiVersion: postgresql.cnpg.io/v1
kind: Cluster
metadata:
  name: hexabase-db
  namespace: hexabase-system
spec:
  instances: 3
  primaryUpdateStrategy: unsupervised
  
  postgresql:
    parameters:
      max_connections: "200"
      shared_buffers: "4GB"
      effective_cache_size: "12GB"
      maintenance_work_mem: "1GB"
      checkpoint_completion_target: "0.9"
      wal_buffers: "16MB"
      default_statistics_target: "100"
      random_page_cost: "1.1"
      effective_io_concurrency: "200"
      work_mem: "20MB"
      min_wal_size: "1GB"
      max_wal_size: "4GB"
  
  bootstrap:
    initdb:
      database: hexabase
      owner: hexabase
      secret:
        name: hexabase-db-auth
  
  storage:
    size: 1Ti
    storageClass: fast-ssd
  
  monitoring:
    enabled: true
  
  backup:
    retentionPolicy: "30d"
    barmanObjectStore:
      destinationPath: "s3://hexabase-backups/postgres"
      s3Credentials:
        accessKeyId:
          name: s3-credentials
          key: ACCESS_KEY_ID
        secretAccessKey:
          name: s3-credentials
          key: SECRET_ACCESS_KEY
EOF
```

### 5. Deploy Redis Cluster

```bash
helm repo add bitnami https://charts.bitnami.com/bitnami

helm install redis bitnami/redis \
  --namespace hexabase-system \
  --set auth.enabled=true \
  --set auth.existingSecret=redis-auth \
  --set auth.existingSecretPasswordKey=password \
  --set sentinel.enabled=true \
  --set sentinel.masterSet=hexabase \
  --set replica.replicaCount=3 \
  --set master.persistence.size=50Gi \
  --set replica.persistence.size=50Gi \
  --set metrics.enabled=true
```

### 6. Deploy NATS

```bash
helm repo add nats https://nats-io.github.io/k8s/helm/charts/

helm install nats nats/nats \
  --namespace hexabase-system \
  --set nats.jetstream.enabled=true \
  --set nats.jetstream.memStorage.size=2Gi \
  --set nats.jetstream.fileStorage.size=50Gi \
  --set cluster.enabled=true \
  --set cluster.replicas=3 \
  --set natsbox.enabled=true \
  --set metrics.enabled=true
```

### 7. Deploy Hexabase Control Plane

#### Create Namespace and Secrets

```bash
kubectl create namespace hexabase-system

# Create database secret
kubectl create secret generic hexabase-db-auth \
  --namespace hexabase-system \
  --from-literal=username=hexabase \
  --from-literal=password=$(openssl rand -base64 32)

# Create JWT keys
openssl genrsa -out jwt-private.pem 4096
openssl rsa -in jwt-private.pem -pubout -out jwt-public.pem

kubectl create secret generic jwt-keys \
  --namespace hexabase-system \
  --from-file=private.pem=jwt-private.pem \
  --from-file=public.pem=jwt-public.pem

# Create OAuth secrets
kubectl create secret generic oauth-providers \
  --namespace hexabase-system \
  --from-literal=google-client-id=$GOOGLE_CLIENT_ID \
  --from-literal=google-client-secret=$GOOGLE_CLIENT_SECRET \
  --from-literal=github-client-id=$GITHUB_CLIENT_ID \
  --from-literal=github-client-secret=$GITHUB_CLIENT_SECRET
```

#### Deploy Hexabase API

```yaml
# hexabase-api.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: hexabase-config
  namespace: hexabase-system
data:
  config.yaml: |
    server:
      port: 8080
      mode: production
    
    database:
      host: hexabase-db-rw
      port: 5432
      name: hexabase
      sslMode: require
      maxOpenConns: 100
      maxIdleConns: 10
      connMaxLifetime: 1h
    
    redis:
      addr: redis-master:6379
      db: 0
      poolSize: 100
    
    nats:
      url: nats://nats:4222
      streamName: hexabase
    
    auth:
      jwtExpiry: 24h
      refreshExpiry: 168h
    
    kubernetes:
      inCluster: true
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hexabase-api
  namespace: hexabase-system
spec:
  replicas: 3
  selector:
    matchLabels:
      app: hexabase-api
  template:
    metadata:
      labels:
        app: hexabase-api
    spec:
      serviceAccountName: hexabase-api
      containers:
      - name: api
        image: hexabase/api:latest
        ports:
        - containerPort: 8080
        env:
        - name: CONFIG_PATH
          value: /etc/hexabase/config.yaml
        - name: DATABASE_PASSWORD
          valueFrom:
            secretKeyRef:
              name: hexabase-db-auth
              key: password
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: redis-auth
              key: password
        volumeMounts:
        - name: config
          mountPath: /etc/hexabase
        - name: jwt-keys
          mountPath: /etc/hexabase/keys
        resources:
          requests:
            cpu: 500m
            memory: 512Mi
          limits:
            cpu: 2000m
            memory: 2Gi
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: config
        configMap:
          name: hexabase-config
      - name: jwt-keys
        secret:
          secretName: jwt-keys
---
apiVersion: v1
kind: Service
metadata:
  name: hexabase-api
  namespace: hexabase-system
spec:
  selector:
    app: hexabase-api
  ports:
  - port: 80
    targetPort: 8080
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: hexabase-api
  namespace: hexabase-system
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/rate-limit: "100"
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - api.hexabase.ai
    secretName: api-tls
  rules:
  - host: api.hexabase.ai
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: hexabase-api
            port:
              number: 80
```

Apply the configuration:

```bash
kubectl apply -f hexabase-api.yaml
```

### 8. Configure Monitoring

#### Install Prometheus Stack

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts

helm install kube-prometheus-stack prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace \
  --set prometheus.prometheusSpec.retention=30d \
  --set prometheus.prometheusSpec.storageSpec.volumeClaimTemplate.spec.resources.requests.storage=100Gi \
  --set alertmanager.alertmanagerSpec.storage.volumeClaimTemplate.spec.resources.requests.storage=10Gi
```

#### Install Loki Stack

```bash
helm repo add grafana https://grafana.github.io/helm-charts

helm install loki grafana/loki-stack \
  --namespace monitoring \
  --set loki.persistence.enabled=true \
  --set loki.persistence.size=100Gi \
  --set promtail.enabled=true
```

### 9. Configure Backup

#### Set Up Velero

```bash
# Install Velero CLI
wget https://github.com/vmware-tanzu/velero/releases/download/v1.13.0/velero-v1.13.0-linux-amd64.tar.gz
tar -xvf velero-v1.13.0-linux-amd64.tar.gz
sudo mv velero-v1.13.0-linux-amd64/velero /usr/local/bin/

# Install Velero in cluster
velero install \
  --provider aws \
  --plugins velero/velero-plugin-for-aws:v1.9.0 \
  --bucket hexabase-velero-backups \
  --secret-file ./credentials-velero \
  --backup-location-config region=us-east-1 \
  --snapshot-location-config region=us-east-1
```

#### Create Backup Schedule

```bash
# Daily backup of all namespaces
velero schedule create daily-backup \
  --schedule="0 2 * * *" \
  --include-namespaces hexabase-system,hexabase-workspaces \
  --ttl 720h
```

### 10. Security Hardening

#### Network Policies

```yaml
# default-deny-all.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny-all
  namespace: hexabase-system
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
---
# allow-api-ingress.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-api-ingress
  namespace: hexabase-system
spec:
  podSelector:
    matchLabels:
      app: hexabase-api
  policyTypes:
  - Ingress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8080
```

#### Pod Security Standards

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: hexabase-system
  labels:
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/audit: restricted
    pod-security.kubernetes.io/warn: restricted
```

## Post-Installation

### 1. Verify Installation

```bash
# Check all pods are running
kubectl get pods -n hexabase-system

# Check API health
curl https://api.hexabase.ai/health

# Check database connection
kubectl exec -n hexabase-system hexabase-db-1 -- psql -U hexabase -c "SELECT version();"
```

### 2. Configure DNS

Point your domain to the load balancer:

```bash
# Get load balancer IP
kubectl get svc -n ingress-nginx ingress-nginx-controller

# Configure DNS A records
api.hexabase.ai -> <LB_IP>
app.hexabase.ai -> <LB_IP>
*.workspaces.hexabase.ai -> <LB_IP>
```

### 3. Initial Admin Setup

```bash
# Create admin user
kubectl exec -n hexabase-system deploy/hexabase-api -- \
  hexabase-cli user create \
    --email admin@hexabase.ai \
    --role super_admin
```

### 4. Configure Observability

Access Grafana:
```bash
kubectl port-forward -n monitoring svc/kube-prometheus-stack-grafana 3000:80
# Default credentials: admin/prom-operator
```

Import Hexabase dashboards:
- API Performance Dashboard
- Workspace Usage Dashboard
- Resource Utilization Dashboard

## Maintenance

### Regular Tasks

1. **Daily**
   - Check backup completion
   - Review error logs
   - Monitor resource usage

2. **Weekly**
   - Review security alerts
   - Check for updates
   - Analyze performance metrics

3. **Monthly**
   - Rotate secrets
   - Update dependencies
   - Capacity planning review

### Upgrade Process

```bash
# 1. Backup current state
velero backup create pre-upgrade-$(date +%Y%m%d)

# 2. Update Helm values
helm upgrade hexabase-api ./charts/hexabase-api \
  --namespace hexabase-system \
  --values production-values.yaml

# 3. Verify upgrade
kubectl rollout status deployment/hexabase-api -n hexabase-system
```

## Troubleshooting

### Common Issues

**API pods not starting**
```bash
kubectl describe pod -n hexabase-system <pod-name>
kubectl logs -n hexabase-system <pod-name> --previous
```

**Database connection issues**
```bash
# Check database cluster status
kubectl get cluster -n hexabase-system
kubectl describe cluster hexabase-db -n hexabase-system
```

**High memory usage**
```bash
# Check resource usage
kubectl top nodes
kubectl top pods -n hexabase-system

# Adjust resource limits if needed
```

## Security Checklist

- [ ] All secrets stored in Kubernetes secrets
- [ ] Network policies configured
- [ ] Pod security standards enforced
- [ ] Regular security scanning enabled
- [ ] Audit logging configured
- [ ] Backup encryption enabled
- [ ] TLS certificates valid and auto-renewing
- [ ] RBAC properly configured
- [ ] Resource quotas set
- [ ] Monitoring alerts configured

## Support

For production support:
- Email: support@hexabase.ai
- Enterprise Support Portal: https://support.hexabase.ai
- 24/7 Hotline: +1-xxx-xxx-xxxx (Enterprise only)# Backup and Recovery Guide

Comprehensive guide for implementing backup and disaster recovery for Hexabase AI platform.

## Overview

The Hexabase AI backup strategy covers:
- **Application Data**: PostgreSQL databases, Redis cache
- **Kubernetes Resources**: Workspaces, configurations, secrets
- **Persistent Volumes**: User data, logs, metrics
- **Object Storage**: Uploaded files, artifacts
- **Disaster Recovery**: Full platform restoration procedures

## Backup Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   PostgreSQL    │────▶│     Velero      │────▶│  Object Storage │
│   Databases     │     │  (K8s Backup)   │     │   (S3/Minio)   │
└─────────────────┘     └─────────────────┘     └─────────────────┘
         │                       │                         │
         │              ┌─────────────────┐               │
         └─────────────▶│    WAL-G/PGX    │───────────────┘
                        │  (DB Streaming)  │
                        └─────────────────┘
```

## Backup Components

### 1. Database Backups

#### PostgreSQL with CloudNativePG

```yaml
# postgres-backup-config.yaml
apiVersion: v1
kind: Secret
metadata:
  name: backup-credentials
  namespace: hexabase-system
stringData:
  ACCESS_KEY_ID: "your-s3-access-key"
  SECRET_ACCESS_KEY: "your-s3-secret-key"
---
apiVersion: postgresql.cnpg.io/v1
kind: Cluster
metadata:
  name: hexabase-db
  namespace: hexabase-system
spec:
  instances: 3
  
  postgresql:
    parameters:
      archive_mode: "on"
      archive_timeout: "5min"
      max_wal_size: "4GB"
      min_wal_size: "1GB"
  
  backup:
    # Retention policy
    retentionPolicy: "30d"
    
    # S3-compatible object store
    barmanObjectStore:
      destinationPath: "s3://hexabase-backups/postgres"
      endpointURL: "https://s3.amazonaws.com"
      s3Credentials:
        accessKeyId:
          name: backup-credentials
          key: ACCESS_KEY_ID
        secretAccessKey:
          name: backup-credentials
          key: SECRET_ACCESS_KEY
      
      # WAL archive configuration
      wal:
        compression: gzip
        encryption: AES256
        maxParallel: 8
      
      # Base backup configuration
      data:
        compression: gzip
        encryption: AES256
        immediateCheckpoint: false
        jobs: 4
```

#### Manual PostgreSQL Backup

```bash
#!/bin/bash
# backup-postgres.sh

# Variables
DB_HOST="hexabase-db-rw.hexabase-system.svc.cluster.local"
DB_NAME="hexabase"
DB_USER="hexabase"
BACKUP_DIR="/backups/postgres"
S3_BUCKET="s3://hexabase-backups/postgres/manual"
DATE=$(date +%Y%m%d_%H%M%S)

# Create backup
kubectl exec -n hexabase-system hexabase-db-1 -- \
  pg_dump -h $DB_HOST -U $DB_USER -d $DB_NAME \
  --format=custom \
  --verbose \
  --no-password \
  --compress=9 \
  > ${BACKUP_DIR}/hexabase_${DATE}.dump

# Encrypt backup
openssl enc -aes-256-cbc -salt \
  -in ${BACKUP_DIR}/hexabase_${DATE}.dump \
  -out ${BACKUP_DIR}/hexabase_${DATE}.dump.enc \
  -pass file:/etc/backup/encryption.key

# Upload to S3
aws s3 cp ${BACKUP_DIR}/hexabase_${DATE}.dump.enc \
  ${S3_BUCKET}/hexabase_${DATE}.dump.enc \
  --storage-class GLACIER_IR

# Cleanup old local backups
find ${BACKUP_DIR} -name "*.dump*" -mtime +7 -delete
```

### 2. Kubernetes Resource Backups

#### Install Velero

```bash
# Download Velero CLI
wget https://github.com/vmware-tanzu/velero/releases/download/v1.13.0/velero-v1.13.0-linux-amd64.tar.gz
tar -xvf velero-v1.13.0-linux-amd64.tar.gz
sudo mv velero-v1.13.0-linux-amd64/velero /usr/local/bin/

# Create S3 credentials
cat > credentials-velero <<EOF
[default]
aws_access_key_id=your-access-key
aws_secret_access_key=your-secret-key
EOF

# Install Velero in cluster
velero install \
  --provider aws \
  --plugins velero/velero-plugin-for-aws:v1.9.0 \
  --bucket hexabase-velero-backups \
  --secret-file ./credentials-velero \
  --backup-location-config \
    region=us-east-1,s3ForcePathStyle=false,s3Url=https://s3.amazonaws.com \
  --snapshot-location-config \
    region=us-east-1 \
  --use-node-agent \
  --default-volumes-to-fs-backup
```

#### Configure Backup Schedules

```bash
# Daily backup of control plane
velero schedule create control-plane-daily \
  --schedule="0 2 * * *" \
  --include-namespaces hexabase-system,hexabase-workspaces \
  --exclude-resources pods,events \
  --ttl 720h \
  --storage-location default

# Hourly backup of critical configs
velero schedule create configs-hourly \
  --schedule="0 * * * *" \
  --include-resources \
    configmaps,secrets,ingresses,services,deployments,statefulsets \
  --ttl 168h

# Weekly full cluster backup
velero schedule create full-weekly \
  --schedule="0 3 * * 0" \
  --ttl 2160h \
  --exclude-namespaces kube-system,kube-public,kube-node-lease
```

### 3. Persistent Volume Backups

#### Configure Volume Snapshots

```yaml
# volume-snapshot-class.yaml
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshotClass
metadata:
  name: csi-snapclass
driver: ebs.csi.aws.com
deletionPolicy: Retain
parameters:
  type: "gp3"
  encrypted: "true"
---
# Create volume snapshot schedule
apiVersion: batch/v1
kind: CronJob
metadata:
  name: volume-snapshot-scheduler
  namespace: hexabase-system
spec:
  schedule: "0 */6 * * *"  # Every 6 hours
  jobTemplate:
    spec:
      template:
        spec:
          serviceAccountName: volume-snapshot-sa
          containers:
          - name: snapshot-creator
            image: hexabase/volume-snapshot-tool:latest
            command:
            - /bin/sh
            - -c
            - |
              # Get all PVCs
              for pvc in $(kubectl get pvc -A -o jsonpath='{range .items[*]}{.metadata.namespace}{" "}{.metadata.name}{"\n"}{end}'); do
                namespace=$(echo $pvc | awk '{print $1}')
                name=$(echo $pvc | awk '{print $2}')
                
                # Create snapshot
                kubectl apply -f - <<EOF
              apiVersion: snapshot.storage.k8s.io/v1
              kind: VolumeSnapshot
              metadata:
                name: ${name}-$(date +%Y%m%d-%H%M%S)
                namespace: ${namespace}
              spec:
                volumeSnapshotClassName: csi-snapclass
                source:
                  persistentVolumeClaimName: ${name}
              EOF
              done
              
              # Cleanup old snapshots (keep last 5)
              kubectl get volumesnapshot -A --sort-by=.metadata.creationTimestamp \
                | tail -n +6 | awk '{print $1" "$2}' \
                | xargs -n2 sh -c 'kubectl delete volumesnapshot -n $0 $1'
          restartPolicy: OnFailure
```

### 4. Application Data Backup

#### Redis Backup

```yaml
# redis-backup-cronjob.yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: redis-backup
  namespace: hexabase-system
spec:
  schedule: "0 */4 * * *"  # Every 4 hours
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: redis-backup
            image: redis:7-alpine
            command:
            - /bin/sh
            - -c
            - |
              # Create backup
              redis-cli -h redis-master --rdb /backup/dump.rdb
              
              # Compress
              gzip -9 /backup/dump.rdb
              
              # Upload to S3
              aws s3 cp /backup/dump.rdb.gz \
                s3://hexabase-backups/redis/dump-$(date +%Y%m%d-%H%M%S).rdb.gz
            volumeMounts:
            - name: backup
              mountPath: /backup
            - name: aws-credentials
              mountPath: /root/.aws
          volumes:
          - name: backup
            emptyDir: {}
          - name: aws-credentials
            secret:
              secretName: aws-credentials
          restartPolicy: OnFailure
```

#### Object Storage Sync

```yaml
# minio-backup-cronjob.yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: object-storage-backup
  namespace: hexabase-system
spec:
  schedule: "0 1 * * *"  # Daily at 1 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: mc-backup
            image: minio/mc:latest
            command:
            - /bin/sh
            - -c
            - |
              # Configure MinIO client
              mc alias set source https://minio.hexabase.local \
                $SOURCE_ACCESS_KEY $SOURCE_SECRET_KEY
              
              mc alias set backup s3 \
                $BACKUP_ACCESS_KEY $BACKUP_SECRET_KEY
              
              # Mirror buckets with versioning
              for bucket in $(mc ls source | awk '{print $5}'); do
                mc mirror source/$bucket backup/hexabase-backup-$bucket \
                  --overwrite --remove \
                  --exclude "*.tmp" \
                  --encrypt-key "backup/=32byteslongsecretkeymustbegiven"
              done
            env:
            - name: SOURCE_ACCESS_KEY
              valueFrom:
                secretKeyRef:
                  name: minio-credentials
                  key: access-key
            - name: SOURCE_SECRET_KEY
              valueFrom:
                secretKeyRef:
                  name: minio-credentials
                  key: secret-key
            - name: BACKUP_ACCESS_KEY
              valueFrom:
                secretKeyRef:
                  name: backup-s3-credentials
                  key: access-key
            - name: BACKUP_SECRET_KEY
              valueFrom:
                secretKeyRef:
                  name: backup-s3-credentials
                  key: secret-key
          restartPolicy: OnFailure
```

## Disaster Recovery Procedures

### 1. Recovery Planning

#### Recovery Time Objectives (RTO)

| Component | RTO | RPO | Priority |
|-----------|-----|-----|----------|
| Control Plane API | 15 min | 5 min | Critical |
| PostgreSQL Database | 30 min | 5 min | Critical |
| Workspace vClusters | 1 hour | 1 hour | High |
| User Data (PVs) | 2 hours | 6 hours | High |
| Monitoring Stack | 4 hours | 24 hours | Medium |
| Log Data | 8 hours | 24 hours | Low |

#### Disaster Recovery Runbook

```bash
#!/bin/bash
# disaster-recovery.sh

# 1. Verify backup availability
echo "=== Verifying Backups ==="
velero backup get
aws s3 ls s3://hexabase-backups/postgres/ --recursive | tail -10

# 2. Prepare new cluster
echo "=== Preparing Recovery Cluster ==="
# Assumes new K3s cluster is ready

# 3. Install Velero in recovery cluster
velero install \
  --provider aws \
  --plugins velero/velero-plugin-for-aws:v1.9.0 \
  --bucket hexabase-velero-backups \
  --secret-file ./credentials-velero \
  --backup-location-config region=us-east-1 \
  --snapshot-location-config region=us-east-1 \
  --use-node-agent \
  --wait

# 4. Restore control plane
echo "=== Restoring Control Plane ==="
LATEST_BACKUP=$(velero backup get --output json | jq -r '.items[0].metadata.name')
velero restore create control-plane-restore \
  --from-backup $LATEST_BACKUP \
  --include-namespaces hexabase-system \
  --wait

# 5. Restore database
echo "=== Restoring PostgreSQL ==="
kubectl apply -f postgres-cluster.yaml
kubectl wait --for=condition=Ready cluster/hexabase-db -n hexabase-system --timeout=600s

# Restore from backup
kubectl exec -n hexabase-system hexabase-db-1 -- \
  barman-cloud-restore \
    --cloud-provider aws-s3 \
    s3://hexabase-backups/postgres \
    $(kubectl get cluster hexabase-db -n hexabase-system -o jsonpath='{.status.targetPrimary}') \
    latest

# 6. Restore workspaces
echo "=== Restoring Workspaces ==="
velero restore create workspaces-restore \
  --from-backup $LATEST_BACKUP \
  --include-namespaces hexabase-workspaces \
  --wait

# 7. Verify restoration
echo "=== Verifying Restoration ==="
kubectl get pods -n hexabase-system
kubectl get vcluster -A
```

### 2. Database Recovery

#### Point-in-Time Recovery (PITR)

```bash
# Restore PostgreSQL to specific point in time
RECOVERY_TIME="2024-01-15 14:30:00"

# Stop current cluster
kubectl scale cluster hexabase-db -n hexabase-system --replicas=0

# Restore with PITR
kubectl exec -n hexabase-system hexabase-db-recovery -- \
  barman-cloud-restore \
    --cloud-provider aws-s3 \
    --endpoint-url https://s3.amazonaws.com \
    s3://hexabase-backups/postgres \
    hexabase-db \
    latest \
    --target-time "$RECOVERY_TIME"

# Update cluster to use recovered data
kubectl patch cluster hexabase-db -n hexabase-system --type merge -p \
  '{"spec":{"bootstrap":{"recovery":{"source":"hexabase-db-recovery"}}}}'

# Scale back up
kubectl scale cluster hexabase-db -n hexabase-system --replicas=3
```

### 3. Workspace Recovery

#### Individual Workspace Restoration

```bash
#!/bin/bash
# restore-workspace.sh

WORKSPACE_ID=$1
BACKUP_NAME=$2

# Find workspace resources in backup
velero backup describe $BACKUP_NAME --details | grep $WORKSPACE_ID

# Restore specific workspace
velero restore create workspace-$WORKSPACE_ID-restore \
  --from-backup $BACKUP_NAME \
  --selector "workspace-id=$WORKSPACE_ID" \
  --include-namespaces hexabase-workspaces \
  --wait

# Restore vCluster
kubectl apply -f - <<EOF
apiVersion: vcluster.loft.sh/v1alpha1
kind: VCluster
metadata:
  name: $WORKSPACE_ID
  namespace: hexabase-workspaces
spec:
  # Restored configuration
  restore:
    backup: $BACKUP_NAME
    workspace: $WORKSPACE_ID
EOF

# Wait for vCluster to be ready
kubectl wait --for=condition=Ready vcluster/$WORKSPACE_ID \
  -n hexabase-workspaces --timeout=300s

# Restore persistent volumes
for pvc in $(kubectl get pvc -n hexabase-workspaces -l workspace-id=$WORKSPACE_ID -o name); do
  SNAPSHOT=$(kubectl get volumesnapshot -n hexabase-workspaces \
    -l pvc-name=$(basename $pvc) \
    --sort-by=.metadata.creationTimestamp \
    -o jsonpath='{.items[-1].metadata.name}')
  
  kubectl patch $pvc -n hexabase-workspaces --type merge -p \
    '{"spec":{"dataSource":{"name":"'$SNAPSHOT'","kind":"VolumeSnapshot","apiGroup":"snapshot.storage.k8s.io"}}}'
done
```

### 4. Data Validation

#### Post-Recovery Validation Script

```bash
#!/bin/bash
# validate-recovery.sh

echo "=== Validating Recovery ==="

# 1. Check API health
API_HEALTH=$(curl -s https://api.hexabase.ai/health | jq -r '.status')
if [ "$API_HEALTH" != "healthy" ]; then
  echo "ERROR: API is not healthy"
  exit 1
fi

# 2. Validate database
DB_CHECK=$(kubectl exec -n hexabase-system hexabase-db-1 -- \
  psql -U hexabase -d hexabase -c "SELECT COUNT(*) FROM workspaces;" -t)
echo "Database contains $DB_CHECK workspaces"

# 3. Check vClusters
VCLUSTERS=$(kubectl get vcluster -A --no-headers | wc -l)
echo "Found $VCLUSTERS vClusters"

# 4. Test workspace connectivity
for vcluster in $(kubectl get vcluster -A -o jsonpath='{range .items[*]}{.metadata.namespace}/{.metadata.name} {end}'); do
  ns=$(echo $vcluster | cut -d/ -f1)
  name=$(echo $vcluster | cut -d/ -f2)
  
  kubectl --kubeconfig=/tmp/kubeconfig-$name \
    --context vcluster-$name \
    get nodes > /dev/null 2>&1
  
  if [ $? -eq 0 ]; then
    echo "✓ vCluster $name is accessible"
  else
    echo "✗ vCluster $name is NOT accessible"
  fi
done

# 5. Verify persistent data
kubectl get pv -o json | jq -r '.items[] | select(.status.phase=="Bound") | .metadata.name' | while read pv; do
  echo "Checking PV: $pv"
  # Add specific data validation based on PV content
done

echo "=== Recovery Validation Complete ==="
```

## Backup Monitoring

### 1. Backup Status Dashboard

```yaml
# grafana-backup-dashboard.json
{
  "dashboard": {
    "title": "Backup Status",
    "panels": [
      {
        "title": "Backup Success Rate",
        "targets": [{
          "expr": "rate(velero_backup_success_total[24h]) / rate(velero_backup_attempt_total[24h])"
        }]
      },
      {
        "title": "Last Successful Backups",
        "targets": [{
          "expr": "time() - velero_backup_last_successful_timestamp"
        }]
      },
      {
        "title": "Backup Size Trend",
        "targets": [{
          "expr": "velero_backup_size_bytes"
        }]
      },
      {
        "title": "PostgreSQL WAL Lag",
        "targets": [{
          "expr": "pg_replication_lag_seconds"
        }]
      }
    ]
  }
}
```

### 2. Backup Alerts

```yaml
# backup-alerts.yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: backup-alerts
  namespace: monitoring
spec:
  groups:
  - name: backup
    rules:
    - alert: BackupFailed
      expr: increase(velero_backup_failure_total[1h]) > 0
      for: 5m
      labels:
        severity: critical
        team: platform
      annotations:
        summary: "Backup failed"
        description: "Velero backup {{ $labels.schedule }} has failed"
    
    - alert: BackupDelayed
      expr: time() - velero_backup_last_successful_timestamp > 86400
      for: 1h
      labels:
        severity: warning
        team: platform
      annotations:
        summary: "Backup delayed"
        description: "No successful backup for {{ $labels.schedule }} in 24 hours"
    
    - alert: PostgreSQLReplicationLag
      expr: pg_replication_lag_seconds > 300
      for: 10m
      labels:
        severity: critical
        team: platform
      annotations:
        summary: "PostgreSQL replication lag"
        description: "Replication lag is {{ $value }} seconds"
```

## Backup Testing

### 1. Automated Recovery Testing

```yaml
# backup-test-cronjob.yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: backup-recovery-test
  namespace: hexabase-system
spec:
  schedule: "0 4 * * 0"  # Weekly on Sunday
  jobTemplate:
    spec:
      template:
        spec:
          serviceAccountName: backup-tester
          containers:
          - name: recovery-test
            image: hexabase/backup-tester:latest
            command:
            - /bin/sh
            - -c
            - |
              # Create test namespace
              kubectl create namespace backup-test-$(date +%Y%m%d)
              
              # Restore latest backup to test namespace
              LATEST_BACKUP=$(velero backup get -o json | jq -r '.items[0].metadata.name')
              velero restore create test-restore-$(date +%Y%m%d) \
                --from-backup $LATEST_BACKUP \
                --namespace-mappings hexabase-system:backup-test-$(date +%Y%m%d) \
                --wait
              
              # Validate restoration
              kubectl wait --for=condition=Ready pods \
                -n backup-test-$(date +%Y%m%d) \
                -l app=hexabase-api \
                --timeout=300s
              
              # Run data validation
              kubectl exec -n backup-test-$(date +%Y%m%d) \
                deployment/hexabase-api -- \
                hexabase-cli validate --mode recovery
              
              # Cleanup
              kubectl delete namespace backup-test-$(date +%Y%m%d)
              
              # Report result
              if [ $? -eq 0 ]; then
                echo "Recovery test passed"
                curl -X POST $SLACK_WEBHOOK -d '{"text":"✅ Weekly backup recovery test passed"}'
              else
                echo "Recovery test failed"
                curl -X POST $SLACK_WEBHOOK -d '{"text":"❌ Weekly backup recovery test FAILED"}'
              fi
          restartPolicy: OnFailure
```

### 2. Backup Integrity Verification

```bash
#!/bin/bash
# verify-backup-integrity.sh

# Verify PostgreSQL backup
aws s3 ls s3://hexabase-backups/postgres/ --recursive | tail -5 | while read line; do
  FILE=$(echo $line | awk '{print $4}')
  SIZE=$(echo $line | awk '{print $3}')
  
  # Download and verify
  aws s3 cp s3://hexabase-backups/postgres/$FILE /tmp/
  
  # Check file integrity
  if [[ $FILE == *.enc ]]; then
    openssl enc -aes-256-cbc -d -salt \
      -in /tmp/$(basename $FILE) \
      -out /tmp/$(basename $FILE .enc) \
      -pass file:/etc/backup/encryption.key
  fi
  
  # Verify PostgreSQL dump
  pg_restore --list /tmp/$(basename $FILE .enc) > /dev/null 2>&1
  if [ $? -eq 0 ]; then
    echo "✓ Backup $FILE is valid"
  else
    echo "✗ Backup $FILE is CORRUPT"
  fi
  
  rm -f /tmp/$(basename $FILE)*
done
```

## Best Practices

### 1. Backup Strategy

- **3-2-1 Rule**: 3 copies, 2 different media, 1 offsite
- **Regular Testing**: Weekly automated recovery tests
- **Encryption**: All backups encrypted at rest and in transit
- **Versioning**: Keep multiple versions with retention policy
- **Documentation**: Maintain detailed recovery procedures

### 2. Security

```yaml
# backup-security-policy.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: velero
  namespace: velero
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: velero-backup-role
rules:
- apiGroups: ["*"]
  resources: ["*"]
  verbs: ["get", "list", "watch"]
- apiGroups: [""]
  resources: ["persistentvolumes", "persistentvolumeclaims"]
  verbs: ["get", "list", "watch", "create", "update", "patch"]
---
# Encrypt backup bucket
aws s3api put-bucket-encryption \
  --bucket hexabase-backups \
  --server-side-encryption-configuration '{
    "Rules": [{
      "ApplyServerSideEncryptionByDefault": {
        "SSEAlgorithm": "aws:kms",
        "KMSMasterKeyID": "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012"
      }
    }]
  }'
```

### 3. Cost Optimization

```bash
# Set lifecycle policies for backup storage
aws s3api put-bucket-lifecycle-configuration \
  --bucket hexabase-backups \
  --lifecycle-configuration file://lifecycle.json

# lifecycle.json
{
  "Rules": [
    {
      "ID": "TransitionToGlacier",
      "Status": "Enabled",
      "Transitions": [
        {
          "Days": 30,
          "StorageClass": "GLACIER"
        },
        {
          "Days": 90,
          "StorageClass": "DEEP_ARCHIVE"
        }
      ]
    },
    {
      "ID": "DeleteOldBackups",
      "Status": "Enabled",
      "Expiration": {
        "Days": 365
      }
    }
  ]
}
```

## Troubleshooting

### Common Issues

**Velero backup stuck in "InProgress"**
```bash
# Check backup logs
velero backup logs <backup-name>

# Check node agent logs
kubectl logs -n velero -l name=node-agent

# Force completion
velero backup delete <backup-name> --confirm
```

**PostgreSQL WAL archiving failing**
```bash
# Check archive status
kubectl exec -n hexabase-system hexabase-db-1 -- \
  psql -U postgres -c "SELECT * FROM pg_stat_archiver;"

# Check S3 connectivity
kubectl exec -n hexabase-system hexabase-db-1 -- \
  aws s3 ls s3://hexabase-backups/postgres/
```

**Recovery validation failing**
```bash
# Check restored resources
kubectl get all -n <restored-namespace>

# Verify data integrity
kubectl exec -n <restored-namespace> deployment/hexabase-api -- \
  hexabase-cli validate --verbose
```

## Documentation

Maintain the following documentation:

1. **Recovery Runbook**: Step-by-step procedures
2. **Contact List**: Who to call during disasters
3. **Architecture Diagrams**: Current and recovery architectures
4. **Credential Inventory**: Where to find backup credentials
5. **Test Results**: History of recovery tests

## Resources

- [Velero Documentation](https://velero.io/docs/)
- [CloudNativePG Backup Guide](https://cloudnative-pg.io/docs/backup/)
- [Kubernetes Volume Snapshots](https://kubernetes.io/docs/concepts/storage/volume-snapshots/)
- [AWS S3 Glacier](https://aws.amazon.com/glacier/)
- [Disaster Recovery Planning](https://aws.amazon.com/disaster-recovery/)# Deployment Policies

This document defines the deployment policies, procedures, and requirements for each environment in the Hexabase AI platform.

## Environment Overview

| Environment | Purpose | URL | Deployment Method | Approval Required |
|-------------|---------|-----|-------------------|-------------------|
| **Local** | Developer testing | `*.localhost` | Manual/Automated | No |
| **Staging** | Pre-production testing | `*.staging.hexabase.ai` | Automated CI/CD | No |
| **Production** | Live service | `*.hexabase.ai` | Automated CI/CD | Yes |

## Local Environment Policies

### Purpose
- Developer testing and debugging
- Feature development
- Unit and integration testing

### Requirements
- **Infrastructure**: Kind/Minikube cluster or Docker Compose
- **Resources**: Minimal (1-2 GB RAM, 2 CPU cores)
- **Data**: Test data only, no production data allowed
- **Secrets**: Development secrets only

### Deployment Process
```bash
# Automated setup
make setup
make dev

# OR Manual deployment
helm install hexabase-ai ./deployments/helm/hexabase-ai \
  --namespace hexabase-dev \
  --values deployments/helm/values-local.yaml
```

### Policies
1. **No production data** - Never use real customer data
2. **Ephemeral** - Can be destroyed and recreated at any time
3. **Self-contained** - All dependencies run locally
4. **Fast iteration** - Hot reloading enabled
5. **Simplified security** - Basic auth, self-signed certificates

### Configuration
- Internal databases (PostgreSQL, Redis, NATS)
- No persistence for Redis
- Simplified monitoring
- Debug logging enabled
- Mock external services allowed

## Staging Environment Policies

### Purpose
- Integration testing
- Performance testing
- User acceptance testing (UAT)
- Demo environment

### Requirements
- **Infrastructure**: Dedicated Kubernetes cluster
- **Resources**: 50% of production capacity
- **Data**: Anonymized production-like data
- **Secrets**: Staging-specific secrets (never share with production)

### Deployment Process

#### Automatic Deployment
- Triggered on merge to `develop` branch
- Automated via GitHub Actions/GitLab CI
- No manual approval required

```yaml
# .github/workflows/deploy-staging.yml
on:
  push:
    branches: [develop]
  workflow_dispatch:

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Deploy to Staging
        run: |
          helm upgrade --install hexabase-ai \
            hexabase/hexabase-ai \
            --namespace hexabase-staging \
            --values deployments/helm/values-staging.yaml \
            --wait
```

### Policies

1. **Automated Testing**
   - All tests must pass before deployment
   - Minimum 70% code coverage
   - Security scanning required
   - Performance benchmarks must be met

2. **Data Management**
   - Production data must be anonymized
   - PII must be scrubbed
   - Retention: 30 days
   - Weekly database refresh from production (anonymized)

3. **Access Control**
   - Basic auth on all endpoints
   - VPN required for direct access
   - Limited to development team and QA

4. **Rollback Policy**
   - Automatic rollback on deployment failure
   - Keep last 5 releases
   - 1-click rollback capability

5. **Monitoring**
   - Same monitoring stack as production
   - Alerts sent to #staging-alerts Slack channel
   - Lower alert thresholds than production

### Configuration
```yaml
# Key staging configurations
replicas:
  api: 2      # Reduced from production
  ui: 1       # Single instance
resources:
  reduced: true  # 50% of production
security:
  basic_auth: enabled
  debug_endpoints: enabled
monitoring:
  retention: 7d  # Shorter retention
backups:
  frequency: weekly
```

### Testing Requirements
- [ ] Unit tests pass (100%)
- [ ] Integration tests pass (100%)
- [ ] E2E tests pass (critical paths)
- [ ] Security scan clean
- [ ] Performance within 20% of targets

## Production Environment Policies

### Purpose
- Live customer service
- Revenue-generating operations
- SLA commitments

### Requirements
- **Infrastructure**: High-availability Kubernetes cluster
- **Resources**: Full capacity with auto-scaling
- **Data**: Real customer data with encryption
- **Secrets**: Production secrets in HashiCorp Vault or Kubernetes Secrets

### Deployment Process

#### Approval Workflow
1. **Release Creation**
   - Tag release in Git: `v1.2.3`
   - Create release notes
   - Generate changelog

2. **Approval Required**
   - Engineering Manager approval
   - DevOps team review
   - Security team sign-off (for security updates)

3. **Deployment Window**
   - Preferred: Tuesday-Thursday, 10 AM - 3 PM PST
   - Emergency fixes: Anytime with incident declaration

4. **Deployment Steps**
```yaml
# Production deployment via CI/CD
on:
  release:
    types: [published]

jobs:
  deploy-production:
    environment: production
    runs-on: ubuntu-latest
    steps:
      - name: Require Approval
        uses: trstringer/manual-approval@v1
        with:
          approvers: engineering-managers,devops-team
          
      - name: Deploy to Production
        run: |
          helm upgrade --install hexabase-ai \
            hexabase/hexabase-ai \
            --namespace hexabase-system \
            --values deployments/helm/values-production.yaml \
            --atomic \
            --timeout 10m
```

### Policies

1. **Change Management**
   - All changes require PR review (2 approvers minimum)
   - Changes must be tested in staging first
   - Database migrations require separate approval
   - Infrastructure changes require CAB approval

2. **Security Requirements**
   - TLS 1.2+ for all communications
   - Encryption at rest for all data
   - WAF enabled
   - DDoS protection active
   - Regular security audits
   - Penetration testing quarterly

3. **High Availability**
   - Minimum 3 replicas for all services
   - Multi-AZ deployment
   - Database replication
   - Redis Sentinel for HA
   - 99.9% uptime SLA

4. **Backup and Recovery**
   - Daily automated backups
   - Point-in-time recovery capability
   - Backup retention: 30 days daily, 12 months monthly
   - Disaster recovery plan tested quarterly
   - RTO: 4 hours, RPO: 1 hour

5. **Monitoring and Alerting**
   - 24/7 monitoring
   - PagerDuty integration
   - Escalation policies defined
   - Runbooks for all alerts
   - SLO/SLI tracking

6. **Rollback Procedures**
   - Blue-green deployment preferred
   - Instant rollback capability
   - Database migration rollback scripts
   - Maximum rollback time: 15 minutes

### Configuration
```yaml
# Production configurations
replicas:
  api: 3-10      # Auto-scaling enabled
  ui: 2-5        # Auto-scaling enabled
resources:
  guaranteed: true  # Resource guarantees
security:
  waf: enabled
  ddos_protection: enabled
  rate_limiting: strict
monitoring:
  retention: 90d
  high_resolution: true
backups:
  frequency: daily
  replication: cross_region
```

### Pre-deployment Checklist
- [ ] All tests pass in staging
- [ ] Security scan completed
- [ ] Performance benchmarks met
- [ ] Release notes prepared
- [ ] Rollback plan documented
- [ ] Customer communication sent (if needed)
- [ ] On-call engineer assigned
- [ ] Monitoring dashboards ready

### Post-deployment Checklist
- [ ] Health checks passing
- [ ] Metrics within normal range
- [ ] No error rate increase
- [ ] Customer reports verified
- [ ] Performance metrics stable
- [ ] Security scans clean

## Emergency Procedures

### Hotfix Process
1. Declare incident in PagerDuty
2. Create hotfix branch from production
3. Minimal testing in staging
4. Emergency approval (1 senior engineer)
5. Deploy with increased monitoring
6. Post-mortem within 48 hours

### Rollback Triggers
- Error rate > 5%
- Response time > 2x baseline
- Health checks failing
- Customer-impacting bugs
- Security vulnerabilities

## Compliance and Auditing

### Audit Requirements
- All deployments logged with:
  - Who deployed
  - What was deployed
  - When it was deployed
  - Approval trail
  - Configuration changes

### Compliance Checks
- SOC 2 compliance
- GDPR requirements
- HIPAA (if applicable)
- PCI DSS (for payment processing)

### Regular Reviews
- Monthly deployment metrics review
- Quarterly policy review
- Annual security audit
- Continuous compliance monitoring

## Version Management

### Versioning Strategy
- Semantic versioning (MAJOR.MINOR.PATCH)
- Breaking changes require major version bump
- New features require minor version bump
- Bug fixes require patch version bump

### Version Support
- Latest version: Full support
- Previous minor version: Security updates only
- Older versions: Best effort support
- EOL notice: 6 months advance

## Communication

### Deployment Notifications

#### Staging
- Slack: #deployments-staging
- No customer communication

#### Production
- Slack: #deployments-production
- Status page update for major changes
- Email notification for breaking changes
- In-app notification for new features

### Incident Communication
- Status page update within 5 minutes
- Customer email within 30 minutes
- Post-mortem published within 5 days

## Metrics and KPIs

### Deployment Metrics
- Deployment frequency
- Lead time for changes
- Mean time to recovery (MTTR)
- Change failure rate

### Targets
- **Local**: Unlimited deployments
- **Staging**: 10+ deployments per day
- **Production**: 2-5 deployments per week
- **MTTR**: < 30 minutes
- **Change failure rate**: < 5%

## Tools and Technologies

### CI/CD Pipeline
- **Source Control**: Git (GitHub/GitLab)
- **CI/CD**: GitHub Actions / GitLab CI / Tekton
- **Container Registry**: Harbor / DockerHub / ECR
- **Helm Repository**: ChartMuseum / Harbor
- **Secret Management**: HashiCorp Vault / Sealed Secrets

### Deployment Tools
- **Orchestration**: Kubernetes
- **Package Manager**: Helm
- **GitOps**: Flux / ArgoCD
- **Service Mesh**: Istio (optional)

### Monitoring Stack
- **Metrics**: Prometheus + Grafana
- **Logs**: Loki + Grafana
- **Traces**: Jaeger / Tempo
- **Alerts**: AlertManager + PagerDuty

## Policy Enforcement

### Automated Checks
- Pre-commit hooks
- CI/CD pipeline gates
- Admission controllers (OPA/Kyverno)
- Security scanning (Trivy)
- Policy as Code

### Manual Reviews
- Code review requirements
- Architecture review for major changes
- Security review for sensitive changes
- Performance review for critical paths# Kubernetes Deployment Guide

This guide provides detailed instructions for deploying Hexabase AI on Kubernetes or K3s clusters.

## Prerequisites

### Cluster Requirements

- Kubernetes v1.24+ or K3s v1.24+
- RBAC enabled
- Storage class for persistent volumes
- Ingress controller (nginx/traefik)
- cert-manager for TLS certificates

### Required Tools

```bash
# Helm 3 (required for both deployment methods)
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl
sudo mv kubectl /usr/local/bin/

# vcluster CLI
curl -L -o vcluster "https://github.com/loft-sh/vcluster/releases/latest/download/vcluster-linux-amd64"
chmod +x vcluster
sudo mv vcluster /usr/local/bin/
```

## Deployment Method 1: Helm Chart (Recommended)

The easiest and recommended way to deploy Hexabase AI is using our official Helm chart.

### 1. Add Hexabase Helm Repository

```bash
helm repo add hexabase https://charts.hexabase.ai
helm repo update
```

### 2. Configure Your Deployment

We provide pre-configured values files for different environments in the `deployments/helm/` directory:

- **Production**: [`values-production.yaml`](../../deployments/helm/values-production.yaml) - Self-hosted infrastructure with HA
- **Development**: [`values-local.yaml`](../../deployments/helm/values-local.yaml) - Local development setup

For production deployments with self-hosted infrastructure, use the production values file which includes:
- External PostgreSQL, Redis, and NATS configuration
- High availability settings
- Security policies
- Monitoring integration
- Backup configuration

You can copy and customize these files for your specific needs:

```bash
# Copy the production values file
cp deployments/helm/values-production.yaml my-values.yaml

# Edit with your specific configuration
vim my-values.yaml
```

Key configuration sections to update:
- `global.domain` - Your domain name
- Database connection details
- OAuth provider credentials
- Storage class names
- Monitoring endpoints

### 3. Install Hexabase AI

```bash
# Create namespace
kubectl create namespace hexabase-system

# Install with Helm
helm install hexabase-ai hexabase/hexabase-ai \
  --namespace hexabase-system \
  --values values.yaml \
  --wait
```

### 4. Verify Installation

```bash
# Check pod status
kubectl get pods -n hexabase-system

# Check ingress
kubectl get ingress -n hexabase-system

# Check services
kubectl get svc -n hexabase-system

# View logs
kubectl logs -n hexabase-system -l app.kubernetes.io/name=hexabase-api
```

### 5. Access the Application

Once the ingress is configured and DNS is set up:
- UI: https://app.hexabase.ai
- API: https://api.hexabase.ai/health

### Helm Chart Configuration Options

The provided values files in `deployments/helm/` contain comprehensive configuration examples:

- **Self-hosted infrastructure**: See [`values-production.yaml`](../../deployments/helm/values-production.yaml)
- **Cloud-managed services**: Modify the external database sections for RDS, Cloud SQL, ElastiCache, etc.
- **High Availability**: Production values include pod anti-affinity, autoscaling, and PDB configurations
- **Security**: Production values include security contexts, network policies, and RBAC settings

For detailed configuration options, run:
```bash
helm show values hexabase/hexabase-ai > all-values.yaml
```

### Upgrading with Helm

```bash
# Update repo
helm repo update hexabase

# Check for updates
helm list -n hexabase-system

# Upgrade to new version
helm upgrade hexabase-ai hexabase/hexabase-ai \
  --namespace hexabase-system \
  --values values.yaml \
  --wait

# Rollback if needed
helm rollback hexabase-ai 1 -n hexabase-system
```

### Uninstalling

```bash
helm uninstall hexabase-ai -n hexabase-system
kubectl delete namespace hexabase-system
```

---

## Deployment Method 2: Manual Deployment (Alternative)

### 1. Create Namespace and Secrets

```bash
# Create namespace
kubectl create namespace hexabase-system

# Create database secret
kubectl create secret generic hexabase-db \
  --namespace hexabase-system \
  --from-literal=username=hexabase \
  --from-literal=password='<secure-password>' \
  --from-literal=database=hexabase_ai

# Create Redis secret
kubectl create secret generic hexabase-redis \
  --namespace hexabase-system \
  --from-literal=password='<redis-password>'

# Create JWT keys
openssl genrsa -out jwt-private.pem 2048
openssl rsa -in jwt-private.pem -pubout -out jwt-public.pem

kubectl create secret generic hexabase-jwt \
  --namespace hexabase-system \
  --from-file=private.pem=jwt-private.pem \
  --from-file=public.pem=jwt-public.pem

# Create OAuth secrets
kubectl create secret generic hexabase-oauth \
  --namespace hexabase-system \
  --from-literal=google-client-id='<client-id>' \
  --from-literal=google-client-secret='<client-secret>' \
  --from-literal=github-client-id='<client-id>' \
  --from-literal=github-client-secret='<client-secret>'
```

### 2. Install vCluster Operator

```bash
# Add vcluster helm repo
helm repo add loft-sh https://charts.loft.sh
helm repo update

# Install vcluster operator
helm upgrade --install vcluster-operator loft-sh/vcluster-k8s \
  --namespace vcluster-system \
  --create-namespace \
  --set operator.enabled=true
```

### 3. Deploy PostgreSQL (or use external)

For production, use managed PostgreSQL (RDS, Cloud SQL). For testing:

```yaml
# postgres.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: postgres-pvc
  namespace: hexabase-system
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 20Gi
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
  namespace: hexabase-system
spec:
  serviceName: postgres
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
      - name: postgres
        image: postgres:14
        env:
        - name: POSTGRES_DB
          valueFrom:
            secretKeyRef:
              name: hexabase-db
              key: database
        - name: POSTGRES_USER
          valueFrom:
            secretKeyRef:
              name: hexabase-db
              key: username
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: hexabase-db
              key: password
        ports:
        - containerPort: 5432
        volumeMounts:
        - name: postgres-storage
          mountPath: /var/lib/postgresql/data
  volumeClaimTemplates:
  - metadata:
      name: postgres-storage
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 20Gi
---
apiVersion: v1
kind: Service
metadata:
  name: postgres
  namespace: hexabase-system
spec:
  ports:
  - port: 5432
  selector:
    app: postgres
```

### 4. Deploy Redis

```yaml
# redis.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
  namespace: hexabase-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
      - name: redis
        image: redis:7-alpine
        command:
        - redis-server
        - --requirepass
        - $(REDIS_PASSWORD)
        env:
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: hexabase-redis
              key: password
        ports:
        - containerPort: 6379
---
apiVersion: v1
kind: Service
metadata:
  name: redis
  namespace: hexabase-system
spec:
  ports:
  - port: 6379
  selector:
    app: redis
```

### 5. Deploy NATS (Message Queue)

```yaml
# nats.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nats
  namespace: hexabase-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nats
  template:
    metadata:
      labels:
        app: nats
    spec:
      containers:
      - name: nats
        image: nats:2.9-alpine
        command:
        - nats-server
        - --js
        - --sd
        - /data
        ports:
        - containerPort: 4222
        - containerPort: 8222
        volumeMounts:
        - name: nats-storage
          mountPath: /data
      volumes:
      - name: nats-storage
        emptyDir: {}
---
apiVersion: v1
kind: Service
metadata:
  name: nats
  namespace: hexabase-system
spec:
  ports:
  - name: client
    port: 4222
  - name: monitoring
    port: 8222
  selector:
    app: nats
```

### 6. Deploy Hexabase AI API

```yaml
# hexabase-api.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: hexabase-config
  namespace: hexabase-system
data:
  config.yaml: |
    server:
      port: 8080
      mode: production
    
    database:
      host: postgres
      port: 5432
      sslmode: require
    
    redis:
      host: redis
      port: 6379
    
    nats:
      url: nats://nats:4222
    
    auth:
      jwt:
        issuer: https://api.hexabase.ai
        audience: hexabase-ai
      oauth:
        redirect_base_url: https://app.hexabase.ai
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hexabase-api
  namespace: hexabase-system
spec:
  replicas: 3
  selector:
    matchLabels:
      app: hexabase-api
  template:
    metadata:
      labels:
        app: hexabase-api
    spec:
      serviceAccountName: hexabase-api
      containers:
      - name: api
        image: hexabase/hexabase-ai-api:latest
        env:
        - name: CONFIG_PATH
          value: /config/config.yaml
        - name: DB_USERNAME
          valueFrom:
            secretKeyRef:
              name: hexabase-db
              key: username
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: hexabase-db
              key: password
        - name: DB_NAME
          valueFrom:
            secretKeyRef:
              name: hexabase-db
              key: database
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: hexabase-redis
              key: password
        - name: GOOGLE_CLIENT_ID
          valueFrom:
            secretKeyRef:
              name: hexabase-oauth
              key: google-client-id
        - name: GOOGLE_CLIENT_SECRET
          valueFrom:
            secretKeyRef:
              name: hexabase-oauth
              key: google-client-secret
        ports:
        - containerPort: 8080
        volumeMounts:
        - name: config
          mountPath: /config
        - name: jwt-keys
          mountPath: /keys
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: config
        configMap:
          name: hexabase-config
      - name: jwt-keys
        secret:
          secretName: hexabase-jwt
---
apiVersion: v1
kind: Service
metadata:
  name: hexabase-api
  namespace: hexabase-system
spec:
  ports:
  - port: 80
    targetPort: 8080
  selector:
    app: hexabase-api
```

### 7. Deploy Frontend

```yaml
# hexabase-ui.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hexabase-ui
  namespace: hexabase-system
spec:
  replicas: 2
  selector:
    matchLabels:
      app: hexabase-ui
  template:
    metadata:
      labels:
        app: hexabase-ui
    spec:
      containers:
      - name: ui
        image: hexabase/hexabase-ai-ui:latest
        env:
        - name: NEXT_PUBLIC_API_URL
          value: https://api.hexabase.ai
        - name: NEXT_PUBLIC_WS_URL
          value: wss://api.hexabase.ai
        ports:
        - containerPort: 3000
---
apiVersion: v1
kind: Service
metadata:
  name: hexabase-ui
  namespace: hexabase-system
spec:
  ports:
  - port: 80
    targetPort: 3000
  selector:
    app: hexabase-ui
```

### 8. Configure Ingress

```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: hexabase-api
  namespace: hexabase-system
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/proxy-body-size: "10m"
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - api.hexabase.ai
    secretName: hexabase-api-tls
  rules:
  - host: api.hexabase.ai
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: hexabase-api
            port:
              number: 80
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: hexabase-ui
  namespace: hexabase-system
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - app.hexabase.ai
    secretName: hexabase-ui-tls
  rules:
  - host: app.hexabase.ai
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: hexabase-ui
            port:
              number: 80
```

### 9. Create RBAC Resources

```yaml
# rbac.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: hexabase-api
  namespace: hexabase-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: hexabase-api
rules:
- apiGroups: [""]
  resources: ["namespaces"]
  verbs: ["get", "list", "create", "delete"]
- apiGroups: ["storage.k8s.io"]
  resources: ["storageclasses"]
  verbs: ["get", "list"]
- apiGroups: ["vcluster.loft.sh"]
  resources: ["vclusters"]
  verbs: ["get", "list", "create", "update", "delete"]
- apiGroups: [""]
  resources: ["secrets", "configmaps"]
  verbs: ["get", "list", "create", "update", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: hexabase-api
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: hexabase-api
subjects:
- kind: ServiceAccount
  name: hexabase-api
  namespace: hexabase-system
```

### 10. Apply All Resources

```bash
# Apply all configurations
kubectl apply -f postgres.yaml
kubectl apply -f redis.yaml
kubectl apply -f nats.yaml
kubectl apply -f rbac.yaml
kubectl apply -f hexabase-api.yaml
kubectl apply -f hexabase-ui.yaml
kubectl apply -f ingress.yaml

# Wait for pods to be ready
kubectl wait --for=condition=ready pod -l app=hexabase-api -n hexabase-system --timeout=300s
kubectl wait --for=condition=ready pod -l app=hexabase-ui -n hexabase-system --timeout=300s
```

### 11. Run Database Migrations

```bash
# Get API pod name
API_POD=$(kubectl get pod -n hexabase-system -l app=hexabase-api -o jsonpath='{.items[0].metadata.name}')

# Run migrations
kubectl exec -n hexabase-system $API_POD -- hexabase-migrate up
```

## Post-Deployment Steps

### 1. Verify Installation

```bash
# Check pod status
kubectl get pods -n hexabase-system

# Check logs
kubectl logs -n hexabase-system -l app=hexabase-api

# Test API health
curl https://api.hexabase.ai/health
```

### 2. Configure DNS

Point your domains to the ingress controller's external IP:
```bash
kubectl get ingress -n hexabase-system
```

### 3. Initial Admin Setup

Access the UI at `https://app.hexabase.ai` and complete initial setup.

## Helm Chart Deployment (Alternative)

For easier deployment, use the Hexabase AI Helm chart:

```bash
# Add Hexabase helm repo
helm repo add hexabase https://charts.hexabase.ai
helm repo update

# Install with custom values
helm install hexabase-ai hexabase/hexabase-ai \
  --namespace hexabase-system \
  --create-namespace \
  --values values.yaml
```

For example configurations, see the values files in [`deployments/helm/`](../../deployments/helm/).

## Deployment Method Comparison

| Aspect | Helm Chart | Manual Deployment |
|--------|------------|-------------------|
| **Ease of Use** | ⭐⭐⭐⭐⭐ Simple one-command deployment | ⭐⭐ Complex, multiple steps |
| **Maintainability** | ⭐⭐⭐⭐⭐ Easy upgrades and rollbacks | ⭐⭐ Manual tracking required |
| **Configuration** | ⭐⭐⭐⭐⭐ Centralized values.yaml | ⭐⭐ Scattered across files |
| **Customization** | ⭐⭐⭐⭐ Extensive options | ⭐⭐⭐⭐⭐ Full control |
| **Production Ready** | ⭐⭐⭐⭐⭐ Best practices built-in | ⭐⭐⭐ Requires expertise |
| **Time to Deploy** | ~5 minutes | ~30-60 minutes |

**Recommendation**: Use the Helm chart for all deployments unless you have specific requirements that necessitate manual deployment.

## Troubleshooting

### Common Helm Issues

**Chart not found:**
```bash
helm search repo hexabase
# If empty, re-add the repo:
helm repo add hexabase https://charts.hexabase.ai
helm repo update
```

**Values validation errors:**
```bash
# Validate your values file
helm lint hexabase/hexabase-ai -f values.yaml

# See all available options
helm show values hexabase/hexabase-ai
```

**Installation timeouts:**
```bash
# Increase timeout
helm install hexabase-ai hexabase/hexabase-ai \
  --timeout 10m \
  --wait
```

### Getting Help

1. **Check Helm release status:**
   ```bash
   helm status hexabase-ai -n hexabase-system
   ```

2. **View rendered manifests:**
   ```bash
   helm get manifest hexabase-ai -n hexabase-system
   ```

3. **Enable debug output:**
   ```bash
   helm install hexabase-ai hexabase/hexabase-ai \
     --debug \
     --dry-run
   ```

## Next Steps

- Set up [Monitoring & Observability](./monitoring-setup.md)
- Configure [Backup & Recovery](./backup-recovery.md)
- Review [Production Setup](./production-setup.md) for hardening
- Explore [Helm Chart Advanced Configuration](https://charts.hexabase.ai/docs)# Deployment Guides

This directory contains guides for deploying and operating the Hexabase AI platform.

## 📚 Available Guides

### [Kubernetes Deployment](./kubernetes-deployment.md)
Basic deployment guide for Kubernetes environments, covering:
- Prerequisites
- Helm chart installation
- Configuration options
- Basic troubleshooting

### [Production Setup](./production-setup.md)
Comprehensive production deployment guide, including:
- Infrastructure requirements and sizing
- High availability configuration
- K3s cluster setup
- Core component installation
- Security hardening
- Post-installation procedures

### [Monitoring Setup](./monitoring-setup.md)
Complete monitoring and observability setup:
- Prometheus and Grafana installation
- Log aggregation with Loki
- ClickHouse for long-term storage
- Custom dashboards and alerts
- SLI/SLO monitoring

### [Backup & Recovery](./backup-recovery.md)
Backup and disaster recovery procedures:
- Backup strategy and components
- Database backup configuration
- Kubernetes resource backups
- Disaster recovery procedures
- Recovery testing automation

### [Deployment Policies](./deployment-policies.md)
Organizational policies and procedures for deployments:
- Deployment approval process
- Change management
- Rollback procedures
- Maintenance windows

## 🚀 Getting Started

If you're setting up Hexabase AI for the first time:

1. **Development/Testing**: Start with [Kubernetes Deployment](./kubernetes-deployment.md)
2. **Production**: Follow [Production Setup](./production-setup.md)
3. **Monitoring**: Implement [Monitoring Setup](./monitoring-setup.md)
4. **Backup**: Configure [Backup & Recovery](./backup-recovery.md)
5. **Policies**: Review [Deployment Policies](./deployment-policies.md)

## 📋 Deployment Checklist

### Pre-Deployment
- [ ] Infrastructure provisioned and validated
- [ ] Network connectivity verified
- [ ] Storage classes configured
- [ ] DNS entries prepared
- [ ] SSL certificates ready
- [ ] Secrets and credentials generated

### Deployment
- [ ] K3s/Kubernetes cluster installed
- [ ] Core components deployed
- [ ] Database cluster operational
- [ ] Ingress controllers configured
- [ ] Initial admin user created

### Post-Deployment
- [ ] Monitoring stack operational
- [ ] Backup schedules configured
- [ ] Security policies applied
- [ ] Documentation updated
- [ ] Team access configured

## 🏗️ Architecture Overview

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Load Balancer │────▶│  Ingress Nginx  │────▶│   Hexabase API  │
└─────────────────┘     └─────────────────┘     └─────────────────┘
                                                          │
                        ┌─────────────────────────────────┴───────────┐
                        │                                             │
                ┌───────▼────────┐  ┌─────────────┐  ┌──────────────▼──┐
                │   PostgreSQL   │  │    Redis    │  │      NATS       │
                │    Cluster     │  │   Cluster   │  │   Message Bus   │
                └────────────────┘  └─────────────┘  └─────────────────┘
```

## 🔧 Common Operations

### Scaling

```bash
# Scale API replicas
kubectl scale deployment hexabase-api -n hexabase-system --replicas=5

# Scale worker nodes
# Add nodes to K3s cluster, they auto-join
```

### Upgrades

```bash
# Backup before upgrade
velero backup create pre-upgrade-$(date +%Y%m%d)

# Upgrade using Helm
helm upgrade hexabase ./charts/hexabase -n hexabase-system
```

### Troubleshooting

```bash
# Check pod status
kubectl get pods -n hexabase-system

# View logs
kubectl logs -n hexabase-system deployment/hexabase-api

# Access metrics
kubectl port-forward -n monitoring svc/grafana 3000:80
```

## 📊 Resource Requirements

### Minimum (Development)
- 3 nodes (2 CPU, 4GB RAM each)
- 100GB storage
- 10Mbps network

### Recommended (Production)
- Control Plane: 3 nodes (8 CPU, 32GB RAM each)
- Workers: 5+ nodes (16 CPU, 64GB RAM each)
- 1TB+ fast SSD storage
- 1Gbps+ network
- Dedicated database nodes

### Enterprise (Large Scale)
- Control Plane: 5 nodes (16 CPU, 64GB RAM each)
- Workers: 20+ nodes (32 CPU, 128GB RAM each)
- Multi-region deployment
- Dedicated infrastructure for each component

## 🔐 Security Considerations

1. **Network Security**
   - Private networks for internal communication
   - Firewall rules properly configured
   - Network policies enforced

2. **Access Control**
   - RBAC configured
   - Service accounts with minimal permissions
   - Regular credential rotation

3. **Data Protection**
   - Encryption at rest
   - Encryption in transit
   - Regular security scanning

4. **Compliance**
   - Audit logging enabled
   - Backup encryption
   - Data residency requirements met

## 📝 Maintenance

### Daily Tasks
- Monitor system health
- Check backup status
- Review error logs

### Weekly Tasks
- Apply security updates
- Review resource usage
- Test disaster recovery

### Monthly Tasks
- Update dependencies
- Capacity planning review
- Security audit

## 🆘 Support

- **Documentation**: Check guides in this directory
- **Community**: [Discord](https://discord.gg/hexabase)
- **Enterprise Support**: support@hexabase.ai
- **Emergency**: +1-xxx-xxx-xxxx (Enterprise only)

## 🔗 Related Documentation

- [Architecture Overview](../../architecture/system-architecture.md)
- [Development Setup](../development/dev-environment-setup.md)
- [API Reference](../../api-reference/README.md)
- [Roadmap](../../roadmap/README.md)# Monitoring Setup Guide

Comprehensive guide for setting up monitoring and observability for Hexabase AI platform.

## Overview

The Hexabase AI monitoring stack includes:
- **Prometheus**: Metrics collection and storage
- **Grafana**: Visualization and dashboards
- **Loki**: Log aggregation
- **Alertmanager**: Alert routing and management
- **ClickHouse**: Long-term log storage
- **OpenTelemetry**: Distributed tracing

## Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Applications  │────▶│   Prometheus    │────▶│     Grafana     │
│  (metrics/logs) │     │  (time-series)  │     │  (visualization)│
└─────────────────┘     └─────────────────┘     └─────────────────┘
         │                                                │
         │              ┌─────────────────┐               │
         └─────────────▶│      Loki       │───────────────┘
                        │  (log storage)  │
                        └─────────────────┘
                                 │
                        ┌─────────────────┐
                        │   ClickHouse    │
                        │ (long-term logs)│
                        └─────────────────┘
```

## Prerequisites

- Kubernetes cluster with Hexabase AI installed
- Helm 3.x
- kubectl configured
- Storage class for persistent volumes
- DNS configured for monitoring endpoints

## Installation

### 1. Create Monitoring Namespace

```bash
kubectl create namespace monitoring
```

### 2. Install Prometheus Stack

```bash
# Add Prometheus community Helm repository
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

# Create values file for Prometheus
cat > prometheus-values.yaml <<EOF
prometheus:
  prometheusSpec:
    retention: 30d
    retentionSize: 100GB
    storageSpec:
      volumeClaimTemplate:
        spec:
          storageClassName: fast-ssd
          accessModes: ["ReadWriteOnce"]
          resources:
            requests:
              storage: 200Gi
    
    # Resource limits
    resources:
      requests:
        cpu: 1000m
        memory: 2Gi
      limits:
        cpu: 2000m
        memory: 4Gi
    
    # Additional scrape configs
    additionalScrapeConfigs:
    - job_name: 'hexabase-api'
      kubernetes_sd_configs:
      - role: pod
        namespaces:
          names:
          - hexabase-system
      relabel_configs:
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)
      - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
        action: replace
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: \$1:\$2
        target_label: __address__

alertmanager:
  alertmanagerSpec:
    storage:
      volumeClaimTemplate:
        spec:
          storageClassName: fast-ssd
          accessModes: ["ReadWriteOnce"]
          resources:
            requests:
              storage: 10Gi
  
  config:
    global:
      resolve_timeout: 5m
      slack_api_url: '$SLACK_WEBHOOK_URL'
    
    route:
      group_by: ['alertname', 'cluster', 'service']
      group_wait: 10s
      group_interval: 10s
      repeat_interval: 12h
      receiver: 'default-receiver'
      routes:
      - match:
          severity: critical
        receiver: 'critical-receiver'
        continue: true
      - match:
          severity: warning
        receiver: 'warning-receiver'
    
    receivers:
    - name: 'default-receiver'
      slack_configs:
      - channel: '#alerts'
        title: 'Hexabase Alert'
        text: '{{ range .Alerts }}{{ .Annotations.summary }}\n{{ end }}'
    
    - name: 'critical-receiver'
      slack_configs:
      - channel: '#critical-alerts'
        title: '🚨 CRITICAL: Hexabase Alert'
      pagerduty_configs:
      - service_key: '$PAGERDUTY_SERVICE_KEY'
    
    - name: 'warning-receiver'
      slack_configs:
      - channel: '#alerts'
        title: '⚠️ Warning: Hexabase Alert'

grafana:
  enabled: true
  adminPassword: '$GRAFANA_ADMIN_PASSWORD'
  
  persistence:
    enabled: true
    storageClassName: fast-ssd
    size: 50Gi
  
  ingress:
    enabled: true
    ingressClassName: nginx
    annotations:
      cert-manager.io/cluster-issuer: letsencrypt-prod
    hosts:
    - monitoring.hexabase.ai
    tls:
    - secretName: grafana-tls
      hosts:
      - monitoring.hexabase.ai
  
  datasources:
    datasources.yaml:
      apiVersion: 1
      datasources:
      - name: Prometheus
        type: prometheus
        url: http://prometheus-kube-prometheus-prometheus:9090
        access: proxy
        isDefault: true
      - name: Loki
        type: loki
        url: http://loki:3100
        access: proxy
EOF

# Install Prometheus stack
helm install kube-prometheus-stack prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --values prometheus-values.yaml
```

### 3. Install Loki for Log Aggregation

```bash
# Add Grafana Helm repository
helm repo add grafana https://grafana.github.io/helm-charts

# Create Loki values
cat > loki-values.yaml <<EOF
loki:
  auth_enabled: false
  
  storage:
    type: filesystem
    filesystem:
      chunks_directory: /data/loki/chunks
      rules_directory: /data/loki/rules
  
  persistence:
    enabled: true
    storageClassName: fast-ssd
    size: 100Gi
  
  config:
    table_manager:
      retention_deletes_enabled: true
      retention_period: 168h  # 7 days
    
    limits_config:
      enforce_metric_name: false
      reject_old_samples: true
      reject_old_samples_max_age: 168h
      max_query_length: 0h
      max_streams_per_user: 10000
    
    ingester:
      chunk_idle_period: 30m
      max_chunk_age: 1h
      chunk_target_size: 1572864
      chunk_retain_period: 30s
      max_transfer_retries: 0

promtail:
  enabled: true
  
  config:
    clients:
    - url: http://loki:3100/loki/api/v1/push
    
    positions:
      filename: /tmp/positions.yaml
    
    target_config:
      sync_period: 10s

    pipeline_stages:
    - regex:
        expression: '^(?P<namespace>\S+)\s+(?P<pod>\S+)\s+(?P<container>\S+)\s+(?P<level>\S+)\s+(?P<message>.*)$'
    - labels:
        namespace:
        pod:
        container:
        level:
    - timestamp:
        source: time
        format: RFC3339
    - output:
        source: message
EOF

# Install Loki
helm install loki grafana/loki-stack \
  --namespace monitoring \
  --values loki-values.yaml
```

### 4. Install ClickHouse for Long-term Storage

```bash
# Create ClickHouse configuration
cat > clickhouse-values.yaml <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: clickhouse-config
  namespace: monitoring
data:
  config.xml: |
    <clickhouse>
      <logger>
        <level>information</level>
        <log>/var/log/clickhouse-server/clickhouse-server.log</log>
        <errorlog>/var/log/clickhouse-server/clickhouse-server.err.log</errorlog>
        <size>1000M</size>
        <count>10</count>
      </logger>
      
      <max_connections>4096</max_connections>
      <keep_alive_timeout>3</keep_alive_timeout>
      <max_concurrent_queries>100</max_concurrent_queries>
      
      <profiles>
        <default>
          <max_memory_usage>10000000000</max_memory_usage>
          <load_balancing>random</load_balancing>
        </default>
      </profiles>
      
      <users>
        <default>
          <password_sha256_hex>$CLICKHOUSE_PASSWORD_SHA256</password_sha256_hex>
          <networks>
            <ip>::/0</ip>
          </networks>
          <profile>default</profile>
          <quota>default</quota>
        </default>
      </users>
    </clickhouse>
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: clickhouse
  namespace: monitoring
spec:
  serviceName: clickhouse
  replicas: 3
  selector:
    matchLabels:
      app: clickhouse
  template:
    metadata:
      labels:
        app: clickhouse
    spec:
      containers:
      - name: clickhouse
        image: clickhouse/clickhouse-server:23.8
        ports:
        - containerPort: 8123
          name: http
        - containerPort: 9000
          name: native
        volumeMounts:
        - name: data
          mountPath: /var/lib/clickhouse
        - name: config
          mountPath: /etc/clickhouse-server/config.d
        resources:
          requests:
            cpu: 2000m
            memory: 8Gi
          limits:
            cpu: 4000m
            memory: 16Gi
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: ["ReadWriteOnce"]
      storageClassName: fast-ssd
      resources:
        requests:
          storage: 500Gi
EOF

kubectl apply -f clickhouse-values.yaml

# Create ClickHouse schema for logs
kubectl exec -n monitoring clickhouse-0 -- clickhouse-client --query "
CREATE DATABASE IF NOT EXISTS logs;

CREATE TABLE IF NOT EXISTS logs.hexabase (
    timestamp DateTime64(3),
    level String,
    namespace String,
    pod String,
    container String,
    message String,
    trace_id String,
    span_id String,
    user_id String,
    workspace_id String,
    project_id String,
    method String,
    path String,
    status_code UInt16,
    duration_ms UInt32,
    INDEX idx_timestamp timestamp TYPE minmax GRANULARITY 1,
    INDEX idx_trace_id trace_id TYPE bloom_filter GRANULARITY 1,
    INDEX idx_workspace workspace_id TYPE bloom_filter GRANULARITY 1
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (namespace, pod, timestamp)
TTL timestamp + INTERVAL 90 DAY;
"
```

### 5. Configure Log Forwarding to ClickHouse

```yaml
# fluent-bit-clickhouse.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: fluent-bit-config
  namespace: monitoring
data:
  fluent-bit.conf: |
    [SERVICE]
        Flush         5
        Log_Level     info
        Daemon        off
        Parsers_File  parsers.conf
    
    [INPUT]
        Name              tail
        Path              /var/log/containers/*hexabase*.log
        Parser            docker
        Tag               kube.*
        Refresh_Interval  5
        Mem_Buf_Limit     50MB
        Skip_Long_Lines   On
    
    [FILTER]
        Name                kubernetes
        Match               kube.*
        Kube_URL            https://kubernetes.default.svc:443
        Kube_CA_File        /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
        Kube_Token_File     /var/run/secrets/kubernetes.io/serviceaccount/token
        Merge_Log           On
        K8S-Logging.Parser  On
        K8S-Logging.Exclude On
    
    [OUTPUT]
        Name          http
        Match         *
        Host          clickhouse.monitoring.svc.cluster.local
        Port          8123
        URI           /
        Format        json_lines
        Header        X-ClickHouse-Database logs
        Header        X-ClickHouse-Table hexabase
        Header        X-ClickHouse-Format JSONEachRow
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: fluent-bit
  namespace: monitoring
spec:
  selector:
    matchLabels:
      app: fluent-bit
  template:
    metadata:
      labels:
        app: fluent-bit
    spec:
      serviceAccountName: fluent-bit
      containers:
      - name: fluent-bit
        image: fluent/fluent-bit:2.1
        volumeMounts:
        - name: config
          mountPath: /fluent-bit/etc/
        - name: varlog
          mountPath: /var/log
        - name: varlibdockercontainers
          mountPath: /var/lib/docker/containers
          readOnly: true
      volumes:
      - name: config
        configMap:
          name: fluent-bit-config
      - name: varlog
        hostPath:
          path: /var/log
      - name: varlibdockercontainers
        hostPath:
          path: /var/lib/docker/containers
```

### 6. Set Up OpenTelemetry for Tracing

```bash
# Install OpenTelemetry Collector
helm repo add open-telemetry https://open-telemetry.github.io/opentelemetry-helm-charts

cat > otel-values.yaml <<EOF
mode: deployment

config:
  receivers:
    otlp:
      protocols:
        grpc:
          endpoint: 0.0.0.0:4317
        http:
          endpoint: 0.0.0.0:4318
  
  processors:
    batch:
      timeout: 1s
      send_batch_size: 1024
    
    memory_limiter:
      check_interval: 1s
      limit_mib: 2048
      spike_limit_mib: 512
  
  exporters:
    prometheus:
      endpoint: 0.0.0.0:8889
    
    jaeger:
      endpoint: jaeger-collector.monitoring.svc.cluster.local:14250
      tls:
        insecure: true
    
    logging:
      loglevel: info
  
  service:
    pipelines:
      traces:
        receivers: [otlp]
        processors: [memory_limiter, batch]
        exporters: [jaeger, logging]
      
      metrics:
        receivers: [otlp]
        processors: [memory_limiter, batch]
        exporters: [prometheus]

service:
  type: ClusterIP
  ports:
    otlp-grpc:
      port: 4317
    otlp-http:
      port: 4318
    metrics:
      port: 8889

resources:
  limits:
    cpu: 1000m
    memory: 2Gi
  requests:
    cpu: 200m
    memory: 400Mi
EOF

helm install opentelemetry-collector open-telemetry/opentelemetry-collector \
  --namespace monitoring \
  --values otel-values.yaml
```

## Grafana Dashboards

### 1. Import Hexabase Dashboards

```bash
# Download Hexabase dashboards
curl -O https://raw.githubusercontent.com/hexabase/monitoring/main/dashboards/api-performance.json
curl -O https://raw.githubusercontent.com/hexabase/monitoring/main/dashboards/workspace-usage.json
curl -O https://raw.githubusercontent.com/hexabase/monitoring/main/dashboards/resource-utilization.json

# Import via Grafana API
GRAFANA_URL="https://monitoring.hexabase.ai"
GRAFANA_API_KEY="your-api-key"

for dashboard in *.json; do
  curl -X POST \
    -H "Authorization: Bearer $GRAFANA_API_KEY" \
    -H "Content-Type: application/json" \
    -d @$dashboard \
    "$GRAFANA_URL/api/dashboards/db"
done
```

### 2. Key Dashboards to Create

#### API Performance Dashboard
- Request rate by endpoint
- Response time percentiles (p50, p95, p99)
- Error rate by status code
- Active connections
- Request size distribution

#### Workspace Usage Dashboard
- Active workspaces
- Resource usage per workspace
- vCluster provisioning times
- Workspace creation/deletion trends
- Cost allocation by workspace

#### Infrastructure Dashboard
- Node CPU/Memory usage
- Pod distribution across nodes
- Storage utilization
- Network traffic
- Certificate expiration

## Alert Rules

### 1. Critical Alerts

```yaml
# critical-alerts.yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: hexabase-critical
  namespace: monitoring
spec:
  groups:
  - name: critical
    interval: 30s
    rules:
    - alert: APIDown
      expr: up{job="hexabase-api"} == 0
      for: 1m
      labels:
        severity: critical
        team: platform
      annotations:
        summary: "Hexabase API is down"
        description: "{{ $labels.instance }} API endpoint has been down for more than 1 minute."
    
    - alert: DatabaseDown
      expr: pg_up == 0
      for: 1m
      labels:
        severity: critical
        team: platform
      annotations:
        summary: "PostgreSQL database is down"
        description: "PostgreSQL instance {{ $labels.instance }} is down."
    
    - alert: HighErrorRate
      expr: |
        sum(rate(http_requests_total{status=~"5.."}[5m])) 
        / 
        sum(rate(http_requests_total[5m])) > 0.05
      for: 5m
      labels:
        severity: critical
        team: platform
      annotations:
        summary: "High API error rate"
        description: "Error rate is above 5% for the last 5 minutes."
    
    - alert: CertificateExpiringSoon
      expr: certmanager_certificate_expiration_timestamp_seconds - time() < 7 * 24 * 3600
      for: 1h
      labels:
        severity: critical
        team: platform
      annotations:
        summary: "Certificate expiring soon"
        description: "Certificate {{ $labels.name }} in namespace {{ $labels.namespace }} expires in less than 7 days."
```

### 2. Warning Alerts

```yaml
# warning-alerts.yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: hexabase-warnings
  namespace: monitoring
spec:
  groups:
  - name: warnings
    interval: 1m
    rules:
    - alert: HighMemoryUsage
      expr: |
        (node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes) 
        / node_memory_MemTotal_bytes > 0.85
      for: 10m
      labels:
        severity: warning
        team: platform
      annotations:
        summary: "High memory usage on node"
        description: "Node {{ $labels.instance }} memory usage is above 85%."
    
    - alert: HighDiskUsage
      expr: |
        (node_filesystem_size_bytes - node_filesystem_avail_bytes) 
        / node_filesystem_size_bytes > 0.80
      for: 10m
      labels:
        severity: warning
        team: platform
      annotations:
        summary: "High disk usage"
        description: "Disk usage on {{ $labels.instance }} is above 80%."
    
    - alert: SlowAPIResponse
      expr: |
        histogram_quantile(0.95, 
          sum(rate(http_request_duration_seconds_bucket[5m])) 
          by (le, endpoint)
        ) > 1
      for: 10m
      labels:
        severity: warning
        team: platform
      annotations:
        summary: "Slow API response times"
        description: "95th percentile response time for {{ $labels.endpoint }} is above 1 second."
    
    - alert: PodCrashLooping
      expr: rate(kube_pod_container_status_restarts_total[15m]) > 0
      for: 5m
      labels:
        severity: warning
        team: platform
      annotations:
        summary: "Pod is crash looping"
        description: "Pod {{ $labels.namespace }}/{{ $labels.pod }} is crash looping."
```

## Custom Metrics

### 1. Application Metrics

```go
// internal/observability/metrics.go
package observability

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // API metrics
    RequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "hexabase_api_request_duration_seconds",
            Help: "API request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint", "status"},
    )
    
    ActiveWorkspaces = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "hexabase_active_workspaces",
            Help: "Number of active workspaces",
        },
    )
    
    WorkspaceResources = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "hexabase_workspace_resources",
            Help: "Resource usage by workspace",
        },
        []string{"workspace_id", "resource_type"},
    )
    
    // Business metrics
    WorkspacesCreated = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "hexabase_workspaces_created_total",
            Help: "Total number of workspaces created",
        },
        []string{"plan"},
    )
    
    APICallsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "hexabase_api_calls_total",
            Help: "Total number of API calls",
        },
        []string{"workspace_id", "endpoint"},
    )
)
```

### 2. SLI/SLO Monitoring

```yaml
# slo-rules.yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: hexabase-slo
  namespace: monitoring
spec:
  groups:
  - name: slo
    interval: 30s
    rules:
    # API Availability SLO: 99.9%
    - record: slo:api_availability:ratio
      expr: |
        sum(rate(http_requests_total{status!~"5.."}[5m]))
        /
        sum(rate(http_requests_total[5m]))
    
    - alert: APIAvailabilitySLO
      expr: slo:api_availability:ratio < 0.999
      for: 5m
      labels:
        severity: critical
        slo: true
      annotations:
        summary: "API availability SLO breach"
        description: "API availability is {{ $value | humanizePercentage }}, below 99.9% SLO"
    
    # Latency SLO: 95% of requests < 500ms
    - record: slo:api_latency:ratio
      expr: |
        histogram_quantile(0.95,
          sum(rate(http_request_duration_seconds_bucket{le="0.5"}[5m]))
          by (le)
        )
    
    - alert: APILatencySLO
      expr: slo:api_latency:ratio < 0.95
      for: 5m
      labels:
        severity: warning
        slo: true
      annotations:
        summary: "API latency SLO breach"
        description: "95th percentile latency SLO breach"
```

## Log Analysis Queries

### ClickHouse Queries

```sql
-- Top errors by workspace
SELECT 
    workspace_id,
    level,
    COUNT(*) as error_count,
    groupArray(message)[1:5] as sample_messages
FROM logs.hexabase
WHERE level = 'ERROR'
    AND timestamp > now() - INTERVAL 1 HOUR
GROUP BY workspace_id, level
ORDER BY error_count DESC
LIMIT 10;

-- API performance by endpoint
SELECT 
    path,
    quantile(0.5)(duration_ms) as p50,
    quantile(0.95)(duration_ms) as p95,
    quantile(0.99)(duration_ms) as p99,
    COUNT(*) as requests
FROM logs.hexabase
WHERE timestamp > now() - INTERVAL 1 HOUR
    AND status_code < 500
GROUP BY path
ORDER BY requests DESC;

-- User activity timeline
SELECT 
    toStartOfMinute(timestamp) as minute,
    COUNT(DISTINCT user_id) as unique_users,
    COUNT(*) as total_requests
FROM logs.hexabase
WHERE timestamp > now() - INTERVAL 1 DAY
GROUP BY minute
ORDER BY minute;
```

### Loki LogQL Queries

```logql
# Error logs from API pods
{namespace="hexabase-system", container="api"} |= "ERROR"

# Slow requests (>1s)
{namespace="hexabase-system"} 
  | json 
  | duration_ms > 1000
  | line_format "{{.timestamp}} {{.path}} {{.duration_ms}}ms"

# Authentication failures
{namespace="hexabase-system"} 
  |= "authentication failed"
  | json
  | line_format "{{.timestamp}} user={{.user_email}} ip={{.client_ip}}"

# Workspace provisioning timeline
{namespace="hexabase-system"} 
  |~ "workspace.*provisioning|vcluster.*created"
  | json
  | line_format "{{.timestamp}} {{.workspace_id}} {{.message}}"
```

## Maintenance

### 1. Retention Policies

```bash
# Configure Prometheus retention
kubectl patch prometheus kube-prometheus-stack-prometheus \
  -n monitoring \
  --type merge \
  -p '{"spec":{"retention":"30d","retentionSize":"100GB"}}'

# Configure Loki retention
kubectl patch configmap loki \
  -n monitoring \
  --type merge \
  -p '{"data":{"loki.yaml":"table_manager:\n  retention_period: 168h"}}'
```

### 2. Backup Monitoring Data

```bash
# Backup Prometheus data
kubectl exec -n monitoring prometheus-kube-prometheus-prometheus-0 -- \
  tar czf /tmp/prometheus-backup.tar.gz /prometheus

kubectl cp monitoring/prometheus-kube-prometheus-prometheus-0:/tmp/prometheus-backup.tar.gz \
  ./prometheus-backup-$(date +%Y%m%d).tar.gz

# Backup Grafana dashboards
kubectl exec -n monitoring deployment/kube-prometheus-stack-grafana -- \
  grafana-cli admin export-dashboard --dir=/tmp/dashboards

kubectl cp monitoring/deployment/kube-prometheus-stack-grafana:/tmp/dashboards \
  ./grafana-dashboards-$(date +%Y%m%d)
```

### 3. Performance Tuning

```yaml
# Optimize Prometheus performance
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-optimization
  namespace: monitoring
data:
  prometheus.yaml: |
    global:
      scrape_interval: 30s      # Reduce frequency
      evaluation_interval: 30s
      external_labels:
        cluster: 'production'
        region: 'us-east-1'
    
    # Optimize TSDB
    storage:
      tsdb:
        out_of_order_time_window: 30m
        min_block_duration: 2h
        max_block_duration: 48h
        retention.size: 100GB
```

## Troubleshooting

### Common Issues

**High Prometheus memory usage**
```bash
# Check memory usage
kubectl top pod -n monitoring -l app.kubernetes.io/name=prometheus

# Reduce metric cardinality
kubectl exec -n monitoring prometheus-kube-prometheus-prometheus-0 -- \
  promtool tsdb analyze /prometheus

# Identify high cardinality metrics
curl -s http://localhost:9090/api/v1/label/__name__/values | \
  jq -r '.data[]' | \
  xargs -I {} curl -s "http://localhost:9090/api/v1/query?query=count(count+by(__name__)({__name__=\"{}\"}))" | \
  jq -r '.data.result[0].value[1] // 0' | \
  sort -nr | head -20
```

**Grafana not loading dashboards**
```bash
# Check datasources
kubectl exec -n monitoring deployment/kube-prometheus-stack-grafana -- \
  grafana-cli admin data-sources list

# Restart Grafana
kubectl rollout restart deployment/kube-prometheus-stack-grafana -n monitoring
```

**Missing logs in Loki**
```bash
# Check Promtail status
kubectl logs -n monitoring daemonset/loki-promtail --tail=100

# Verify log parsing
kubectl exec -n monitoring daemonset/loki-promtail -- \
  promtail --dry-run --config.file=/etc/promtail/config.yml
```

## Security Best Practices

1. **Enable authentication** for all monitoring endpoints
2. **Use TLS** for all monitoring traffic
3. **Implement RBAC** for Grafana users
4. **Rotate credentials** regularly
5. **Audit access** to monitoring systems
6. **Encrypt backups** of monitoring data
7. **Restrict network access** to monitoring endpoints

## Resources

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [Loki Documentation](https://grafana.com/docs/loki/)
- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [ClickHouse Documentation](https://clickhouse.com/docs/)