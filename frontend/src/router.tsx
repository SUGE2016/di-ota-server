import { createHashRouter, Navigate, Outlet } from 'react-router-dom';
import { LoginPage } from './pages/LoginPage';
import { MainLayout } from './components/MainLayout';
import { DashboardPage } from './pages/DashboardPage';
import { PackagesPage } from './pages/PackagesPage';
import { PackageDetailPage } from './pages/PackageDetailPage';
import { TasksPage } from './pages/TasksPage';
import { TaskDetailPage } from './pages/TaskDetailPage';
import { UsersPage } from './pages/UsersPage';
import { DevicesPage } from './pages/DevicesPage';
import { AlertsPage } from './pages/AlertsPage';
import { UserDetailPage } from './pages/UserDetailPage';
import { DeviceDetailPage } from './pages/DeviceDetailPage';
import { DeviceSecretsPage } from './pages/DeviceSecretsPage';
import { DeviceSimulatorPage } from './pages/DeviceSimulatorPage';
import useAuthStore from './stores/authStore';
import { canManageDeviceSecrets } from './utils/roles';

function PrivateRoute() {
  const isLoggedIn = useAuthStore((s) => s.isLoggedIn);
  if (!isLoggedIn) {
    return <Navigate to="/login" replace />;
  }
  return <Outlet />;
}

function SecretAdminRoute() {
  const roles = useAuthStore((s) => s.roles);
  if (!canManageDeviceSecrets(roles)) {
    return <Navigate to="/dashboard" replace />;
  }
  return <Outlet />;
}

export const router = createHashRouter([
  { path: '/login', element: <LoginPage /> },
  { path: '/simulator', element: <DeviceSimulatorPage /> },
  {
    element: <PrivateRoute />,
    children: [
      { path: '/', element: <Navigate to="/dashboard" replace /> },
      { element: <MainLayout />, children: [
        { path: '/dashboard', element: <DashboardPage /> },
        { path: '/users', element: <UsersPage /> },
        { path: '/users/:id', element: <UserDetailPage /> },
        { path: '/devices', element: <DevicesPage /> },
        { path: '/devices/:id', element: <DeviceDetailPage /> },
        { element: <SecretAdminRoute />, children: [
          { path: '/device-secrets', element: <DeviceSecretsPage /> },
        ]},
        { path: '/alerts', element: <AlertsPage /> },
        { path: '/packages', element: <PackagesPage /> },
        { path: '/packages/:id', element: <PackageDetailPage /> },
        { path: '/tasks', element: <TasksPage /> },
        { path: '/tasks/:id', element: <TaskDetailPage /> },
      ]},
    ],
  },
]);
