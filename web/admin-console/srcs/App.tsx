import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import { UserManagement } from './pages/UserManagement';
import { ApplicationManagement } from './pages/ApplicationManagement';
import { RoleManagement } from './pages/RoleManagement';
import { AuditLogs } from './pages/AuditLogs';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import './App.css';

const queryClient = new QueryClient();

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <Router>
        <div>
          <nav>
            <ul>
              <li>
                <Link to="/users">User Management</Link>
              </li>
              <li>
                <Link to="/applications">Application Management</Link>
              </li>
              <li>
                <Link to="/roles">Role Management</Link>
              </li>
              <li>
                <Link to="/audit-logs">Audit Logs</Link>
              </li>
            </ul>
          </nav>

          <hr />

          <Routes>
            <Route path="/users" element={<UserManagement />} />
            <Route path="/applications" element={<ApplicationManagement />} />
            <Route path="/roles" element={<RoleManagement />} />
            <Route path="/audit-logs" element={<AuditLogs />} />
          </Routes>
        </div>
      </Router>
    </QueryClientProvider>
  );
}

export default App;
