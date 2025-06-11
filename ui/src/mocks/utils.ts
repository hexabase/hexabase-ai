import { http, HttpResponse } from 'msw';
import { server } from './server';

/**
 * Utility to override specific handlers for testing error scenarios
 */
export const mockApiError = (method: string, path: string, status: number = 500, message: string = 'Internal Server Error') => {
  server.use(
    http[method as keyof typeof http](path, () => {
      return HttpResponse.json(
        { error: message },
        { status }
      );
    })
  );
};

/**
 * Utility to simulate network delay
 */
export const mockApiDelay = (method: string, path: string, delay: number) => {
  server.use(
    http[method as keyof typeof http](path, async () => {
      await new Promise(resolve => setTimeout(resolve, delay));
      return HttpResponse.json({});
    })
  );
};

/**
 * Utility to mock successful empty response
 */
export const mockApiSuccess = (method: string, path: string, data: any = {}) => {
  server.use(
    http[method as keyof typeof http](path, () => {
      return HttpResponse.json(data);
    })
  );
};

/**
 * Reset all handlers to defaults
 */
export const resetMockApi = () => {
  server.resetHandlers();
};