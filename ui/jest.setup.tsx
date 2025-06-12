// Learn more: https://github.com/testing-library/jest-dom
import '@testing-library/jest-dom'
import React from 'react'

// Import mock API client
import { mockApiClient } from './src/test-utils/mock-api-client'

// Mock the API client module
jest.mock('./src/lib/api-client', () => ({
  ...jest.requireActual('./src/lib/api-client'),
  apiClient: mockApiClient,
  authApi: mockApiClient.auth,
  organizationsApi: mockApiClient.organizations,
  workspacesApi: mockApiClient.workspaces,
  projectsApi: mockApiClient.projects,
  applicationsApi: mockApiClient.applications,
  functionsApi: mockApiClient.functions,
}))

// Reset mock function calls between tests
afterEach(() => {
  jest.clearAllMocks()
})

// Mock next/navigation
jest.mock('next/navigation', () => ({
  useRouter() {
    return {
      push: jest.fn(),
      replace: jest.fn(),
      prefetch: jest.fn(),
      back: jest.fn(),
      pathname: '/',
      query: {},
      asPath: '/',
    }
  },
  useSearchParams() {
    return {
      get: jest.fn(),
    }
  },
  usePathname() {
    return '/'
  },
}))

// Mock next/image
const MockedImage = (props: any) => {
  // eslint-disable-next-line @next/next/no-img-element, jsx-a11y/alt-text
  return <img {...props} />
}

jest.mock('next/image', () => ({
  __esModule: true,
  default: MockedImage,
}))

// Mock environment variables
process.env.NEXT_PUBLIC_API_URL = 'http://localhost:8080'
process.env.NEXT_PUBLIC_WS_URL = 'ws://localhost:8080'

// Mock auth context to avoid NextAuth SessionProvider requirement
jest.mock('@/lib/auth-context')

// Mock date-fns functions
jest.mock('date-fns', () => ({
  format: jest.fn((date, formatStr) => {
    if (date instanceof Date) {
      return date.toISOString();
    }
    return String(date);
  }),
  formatDistanceToNow: jest.fn(() => '5 minutes ago'),
  parseISO: jest.fn((date) => new Date(date)),
  isValid: jest.fn(() => true),
  addDays: jest.fn((date, days) => new Date(date)),
  subDays: jest.fn((date, days) => new Date(date)),
}))

// Mock toast hook
jest.mock('@/hooks/use-toast', () => ({
  useToast: () => ({
    toast: jest.fn(),
    toasts: [],
    dismiss: jest.fn(),
  }),
}))

// Mock class-variance-authority
jest.mock('class-variance-authority', () => ({
  cva: () => () => '',
  cx: (...args) => args.filter(Boolean).join(' '),
}))

// Mock @radix-ui/react-alert-dialog
jest.mock('@radix-ui/react-alert-dialog', () => ({
  Root: ({ children, open, onOpenChange, ...props }) => open !== false ? <div data-testid="alert-dialog" {...props}>{children}</div> : null,
  Trigger: ({ children, ...props }) => <button {...props}>{children}</button>,
  Portal: ({ children }) => children,
  Overlay: ({ children, ...props }) => <div {...props}>{children}</div>,
  Content: ({ children, ...props }) => <div {...props}>{children}</div>,
  Title: ({ children, ...props }) => <h2 {...props}>{children}</h2>,
  Description: ({ children, ...props }) => <p {...props}>{children}</p>,
  Action: ({ children, ...props }) => <button {...props}>{children}</button>,
  Cancel: ({ children, ...props }) => <button {...props}>{children}</button>,
}))

// Mock @radix-ui/react-dialog
jest.mock('@radix-ui/react-dialog', () => ({
  Root: ({ children, open, onOpenChange, ...props }) => open !== false ? <div data-testid="dialog" {...props}>{children}</div> : null,
  Trigger: ({ children, ...props }) => <button {...props}>{children}</button>,
  Portal: ({ children }) => children,
  Overlay: ({ children, ...props }) => <div {...props}>{children}</div>,
  Content: ({ children, ...props }) => <div role="dialog" {...props}>{children}</div>,
  Header: ({ children, ...props }) => <div {...props}>{children}</div>,
  Footer: ({ children, ...props }) => <div {...props}>{children}</div>,
  Title: ({ children, ...props }) => <h2 {...props}>{children}</h2>,
  Description: ({ children, ...props }) => <p {...props}>{children}</p>,
  Close: ({ children, ...props }) => <button {...props}>{children}</button>,
}))

// Mock @radix-ui/react-select
jest.mock('@radix-ui/react-select', () => {
  const Select = {
    Root: ({ children, value, onValueChange, ...props }) => {
      // Pass the value and onValueChange to children
      return React.Children.map(children, child => {
        if (React.isValidElement(child) && child.type === Select.Trigger) {
          return React.cloneElement(child, { value, onValueChange } as any);
        }
        return child;
      });
    },
    Trigger: ({ children, id, value, onValueChange, ...props }) => (
      <select 
        id={id}
        value={value} 
        onChange={(e) => onValueChange && onValueChange(e.target.value)}
        {...props}
      >
        <option value="1h">Last 1 hour</option>
        <option value="6h">Last 6 hours</option>
        <option value="24h">Last 24 hours</option>
        <option value="7d">Last 7 days</option>
      </select>
    ),
    Portal: ({ children }) => children,
    Content: ({ children }) => <>{children}</>,
    Item: ({ children, value, ...props }) => (
      <option value={value} {...props}>{children}</option>
    ),
    Value: ({ children, placeholder }) => <>{children || placeholder}</>,
    ItemText: ({ children }) => <>{children}</>,
    ItemIndicator: () => null,
    ScrollUpButton: () => null,
    ScrollDownButton: () => null,
    Viewport: ({ children }) => <>{children}</>,
    Label: ({ children, ...props }) => <label {...props}>{children}</label>,
    Separator: () => <hr />,
    Icon: ({ children, asChild, ...props }) => {
      if (asChild && React.isValidElement(children)) {
        return React.cloneElement(children, props);
      }
      return <span {...props}>{children}</span>;
    },
  };
  
  // Add displayName properties
  Object.keys(Select).forEach(key => {
    if (Select[key]) {
      Select[key].displayName = `Select${key}`;
    }
  });
  
  return Select;
})

// Mock @radix-ui/react-scroll-area
jest.mock('@radix-ui/react-scroll-area', () => {
  const ScrollArea = {
    Root: React.forwardRef(({ children, ...props }, ref) => (
      <div ref={ref} data-testid="scroll-area" {...props}>{children}</div>
    )),
    Viewport: React.forwardRef(({ children, ...props }, ref) => (
      <div ref={ref} {...props}>{children}</div>
    )),
    ScrollAreaScrollbar: React.forwardRef((props, ref) => <div ref={ref} {...props} />),
    ScrollAreaThumb: React.forwardRef((props, ref) => <div ref={ref} {...props} />),
    Corner: React.forwardRef((props, ref) => <div ref={ref} {...props} />),
  };
  
  // Add displayName properties
  Object.keys(ScrollArea).forEach(key => {
    if (ScrollArea[key]) {
      ScrollArea[key].displayName = `ScrollArea${key}`;
    }
  });
  
  return ScrollArea;
})

// Mock the ScrollArea component directly
jest.mock('@/components/ui/scroll-area', () => ({
  ScrollArea: React.forwardRef(({ children, ...props }, ref) => (
    <div ref={ref} data-testid="scroll-area" {...props}>{children}</div>
  )),
  ScrollBar: React.forwardRef((props, ref) => <div ref={ref} {...props} />),
}))

// AIChat component is now implemented - no need to mock

// Mock @radix-ui/react-tooltip
jest.mock('@radix-ui/react-tooltip', () => ({
  Provider: ({ children }) => children,
  Root: ({ children }) => children,
  Trigger: ({ children, ...props }) => React.createElement('button', props, children),
  Portal: ({ children }) => children,
  Content: React.forwardRef(({ children, ...props }, ref) => 
    React.createElement('div', { ref, ...props }, children)
  ),
}))

// Mock lucide-react icons
jest.mock('lucide-react', () => ({
  Plus: () => 'Plus',
  RefreshCw: () => 'RefreshCw',
  Building2: () => 'Building2',
  Loader2: () => 'Loader2',
  X: () => 'X',
  ChevronRight: () => 'ChevronRight',
  ChevronDown: () => 'ChevronDown',
  Check: () => 'Check',
  Server: () => 'Server',
  Database: () => 'Database',
  Calendar: () => 'Calendar',
  Function: () => 'Function',
  Cloud: () => 'Cloud',
  Globe: () => 'Globe',
  Copy: () => 'Copy',
  Edit: () => 'Edit',
  Edit2: () => 'Edit2',
  Trash: () => 'Trash',
  Trash2: () => 'Trash2',
  MoreVertical: () => 'MoreVertical',
  Play: () => 'Play',
  Pause: () => 'Pause',
  History: () => 'History',
  Filter: () => 'Filter',
  Clock: () => 'Clock',
  AlertCircle: () => 'AlertCircle',
  CheckCircle: () => 'CheckCircle',
  XCircle: () => 'XCircle',
  Info: () => 'Info',
  Terminal: () => 'Terminal',
  Settings: () => 'Settings',
  LogOut: () => 'LogOut',
  Home: () => 'Home',
  FolderOpen: () => 'FolderOpen',
  GitBranch: () => 'GitBranch',
  Activity: () => 'Activity',
  Package: () => 'Package',
  Cpu: () => 'Cpu',
  HardDrive: () => 'HardDrive',
  MemoryStick: () => 'MemoryStick',
  Users: () => 'Users',
  Layers: () => 'Layers',
  Shield: () => 'Shield',
  CreditCard: () => 'CreditCard',
  BarChart: () => 'BarChart',
  FileText: () => 'FileText',
  Download: () => 'Download',
  Upload: () => 'Upload',
  Save: () => 'Save',
  Zap: () => 'Zap',
  AlertCircle: () => 'AlertCircle',
  RotateCcw: () => 'RotateCcw',
  DollarSign: () => 'DollarSign',
  ChevronRight: () => 'ChevronRight',
  Loader2: () => 'Loader2',
  Send: () => 'Send',
  Image: () => 'Image',
  FileCode: () => 'FileCode',
}))

// Suppress console errors in tests unless explicitly testing error cases
const originalError = console.error
beforeAll(() => {
  console.error = (...args: any[]) => {
    if (
      typeof args[0] === 'string' &&
      args[0].includes('Warning: ReactDOM.render')
    ) {
      return
    }
    originalError.call(console, ...args)
  }
})

afterAll(() => {
  console.error = originalError
})

// Add custom matchers if needed
expect.extend({
  toBeValidDate(received) {
    const pass = received instanceof Date && !isNaN(received.getTime())
    if (pass) {
      return {
        message: () => `expected ${received} not to be a valid date`,
        pass: true,
      }
    } else {
      return {
        message: () => `expected ${received} to be a valid date`,
        pass: false,
      }
    }
  },
})