import { render, screen } from '@testing-library/react';
import { UserManagement } from '../pages/UserManagement';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { describe, it, expect } from 'vitest';

const queryClient = new QueryClient();

describe('UserManagement', () => {
  it('renders the user management page', () => {
    render(
      <QueryClientProvider client={queryClient}>
        <UserManagement />
      </QueryClientProvider>
    );
    expect(screen.getByText('用户管理')).toBeInTheDocument();
  });
});
