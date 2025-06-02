# Hexabase KaaS Frontend

Next.js frontend application for the Hexabase KaaS platform.

## Features

- 🔐 OAuth authentication with Google & GitHub
- 🏢 Organization management dashboard
- 📱 Responsive design with Tailwind CSS
- ⚡ Built with Next.js 15 and TypeScript
- 🎨 Modern UI components and design system

## Getting Started

### Prerequisites

- Node.js 18+ and npm
- Backend API running on `http://localhost:8080`

### Development Setup

1. **Install dependencies:**
   ```bash
   npm install
   ```

2. **Configure environment:**
   ```bash
   cp .env.local.example .env.local
   # Edit .env.local with your API URL
   ```

3. **Start development server:**
   ```bash
   npm run dev
   ```

4. **Open in browser:**
   ```
   http://localhost:3000
   ```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `NEXT_PUBLIC_API_URL` | Backend API URL | `http://localhost:8080` |

## Project Structure

```
ui/
├── src/
│   ├── app/                 # Next.js App Router pages
│   ├── components/          # React components
│   │   ├── ui/             # Reusable UI components
│   │   ├── login-page.tsx  # OAuth login interface
│   │   └── dashboard-page.tsx # Main dashboard
│   └── lib/                # Utility functions
│       ├── api-client.ts   # API integration
│       ├── auth-context.tsx # Authentication state
│       └── utils.ts        # Helper functions
├── public/                 # Static assets
└── tailwind.config.js     # Tailwind CSS configuration
```

## Available Scripts

- `npm run dev` - Start development server
- `npm run build` - Build for production
- `npm run start` - Start production server
- `npm run lint` - Run ESLint

## Authentication Flow

1. User clicks "Continue with Google/GitHub"
2. Redirect to OAuth provider (Google/GitHub)
3. Provider redirects back with authorization code
4. Backend exchanges code for user info and creates JWT
5. Frontend receives JWT token and stores in cookie
6. User is authenticated and can access dashboard

## API Integration

The frontend communicates with the backend API using:

- **Authentication**: JWT tokens in Authorization header
- **Organizations**: Full CRUD operations
- **Error Handling**: Automatic token refresh and error states
- **Loading States**: UI feedback during API calls

## UI Components

### Core Components

- **LoginPage**: OAuth provider selection
- **DashboardPage**: Main authenticated interface
- **OrganizationsList**: Display and manage organizations
- **CreateOrganizationDialog**: Create new organizations
- **EditOrganizationDialog**: Update organization details

### UI System

- **Button**: Styled button component with variants
- **LoadingSpinner**: Loading indicators
- **Design Tokens**: Consistent spacing, colors, typography

## Development

### Code Style

- TypeScript for type safety
- ESLint for code quality
- Tailwind CSS for styling
- Component-driven architecture

### Testing

```bash
# Run type checking
npm run build

# Run linting
npm run lint
```

## Deployment

### Production Build

```bash
npm run build
npm run start
```

### Environment Setup

Ensure the following environment variables are set:

- `NEXT_PUBLIC_API_URL`: Backend API endpoint
- `NODE_ENV=production`

## Contributing

1. Follow existing code style and patterns
2. Use TypeScript for all new code
3. Ensure responsive design with Tailwind CSS
4. Test authentication flows thoroughly
5. Handle loading and error states properly

## Architecture Decisions

- **Next.js App Router**: Latest routing system
- **Client-Side Authentication**: JWT tokens with cookie storage
- **State Management**: React Context for auth state
- **Styling**: Tailwind CSS for utility-first design
- **API Client**: Axios with interceptors for auth
- **TypeScript**: Full type safety across the application
