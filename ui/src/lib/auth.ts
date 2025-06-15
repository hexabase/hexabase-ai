import { NextAuthOptions } from 'next-auth';
import GoogleProvider from 'next-auth/providers/google';
import GitHubProvider from 'next-auth/providers/github';
import AzureADProvider from 'next-auth/providers/azure-ad';
import CredentialsProvider from 'next-auth/providers/credentials';

export const authOptions: NextAuthOptions = {
  providers: [
    // Development mock provider
    ...(process.env.NODE_ENV === 'development' ? [
      CredentialsProvider({
        id: 'credentials',
        name: 'Development',
        credentials: {
          email: { label: "Email", type: "email" },
          password: { label: "Password", type: "password" }
        },
        async authorize(credentials) {
          // For development, we'll create a mock token that the backend will accept
          if (credentials?.email === "test@hexabase.com" && credentials?.password === "test") {
            // Generate a development token that backend will recognize
            const devToken = 'dev_token_' + Date.now();
            
            // Return user with token info that will be stored in JWT callback
            return {
              id: "dev-user-1",
              name: "Test User",
              email: "test@hexabase.com",
              image: "https://ui-avatars.com/api/?name=Test+User",
              accessToken: devToken,
              refreshToken: 'dev_refresh_' + Date.now(),
              provider: 'credentials'
            };
          }
          return null;
        }
      })
    ] : []),
    // Production providers
    ...(process.env.GOOGLE_CLIENT_ID ? [
      GoogleProvider({
        clientId: process.env.GOOGLE_CLIENT_ID || '',
        clientSecret: process.env.GOOGLE_CLIENT_SECRET || '',
      })
    ] : []),
    ...(process.env.GITHUB_ID ? [
      GitHubProvider({
        clientId: process.env.GITHUB_ID || '',
        clientSecret: process.env.GITHUB_SECRET || '',
      })
    ] : []),
    ...(process.env.AZURE_AD_CLIENT_ID ? [
      AzureADProvider({
        clientId: process.env.AZURE_AD_CLIENT_ID || '',
        clientSecret: process.env.AZURE_AD_CLIENT_SECRET || '',
        tenantId: process.env.AZURE_AD_TENANT_ID || '',
      })
    ] : []),
  ],
  callbacks: {
    async jwt({ token, account, profile, user }) {
      // Initial sign in
      if (account) {
        token.provider = account.provider;
        token.providerId = account.providerAccountId;
        
        // For credentials provider (development), store the tokens
        if (account.provider === 'credentials' && user) {
          token.accessToken = (user as any).accessToken;
          token.refreshToken = (user as any).refreshToken;
        }
      }
      return token;
    },
    async session({ session, token }) {
      if (session.user) {
        // Add custom properties to the user
        (session.user as any).provider = token.provider;
        (session.user as any).providerId = token.providerId;
        
        // For development mode, ensure tokens are available
        if (token.provider === 'credentials') {
          (session as any).accessToken = token.accessToken;
          (session as any).refreshToken = token.refreshToken;
        }
      }
      return session;
    },
    async signIn({ user, account, profile }) {
      // For development credentials, set the token cookies
      if (account?.provider === 'credentials' && typeof window !== 'undefined') {
        // This runs on the client side after successful sign in
        setTimeout(() => {
          if ((user as any).accessToken) {
            // Set cookies that the API client expects
            document.cookie = `hexabase_access_token=${(user as any).accessToken}; path=/; max-age=604800; SameSite=Strict`;
            document.cookie = `hexabase_refresh_token=${(user as any).refreshToken}; path=/; max-age=604800; SameSite=Strict`;
          }
        }, 100);
      }
      return true;
    },
  },
  pages: {
    signIn: '/login',
    error: '/auth/error',
  },
  secret: process.env.NEXTAUTH_SECRET,
};