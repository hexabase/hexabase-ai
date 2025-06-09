# Frontend Development Guide

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
- [Zustand Documentation](https://github.com/pmndrs/zustand)