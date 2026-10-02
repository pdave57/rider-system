import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider, useAuth } from './contexts/AuthContext';
import { Header } from './components/Header';
import { BottomNavigation } from './components/BottomNavigation';

import { Login } from './pages/auth/Login';
import { Register } from './pages/auth/Register';
import { RoleSelect } from './pages/auth/RoleSelect';

import { ClientDashboard } from './pages/client/Dashboard';
import { ClientOrders } from './pages/client/Orders';
import { NewOrder } from './pages/client/NewOrder';
import { ClientTrack } from './pages/client/Track';
import { ProfileSetup } from './pages/client/ProfileSetup';

function ProtectedRoute({ children, role }: { children: JSX.Element; role?: 'client' | 'rider' }) {
  const { user, loading } = useAuth();

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-navy-blue"></div>
      </div>
    );
  }

  if (!user) {
    return <Navigate to="/login" replace />;
  }

  if (role && user.role !== role) {
    return <Navigate to={user.role === 'rider' ? '/rider/dashboard' : '/client/dashboard'} replace />;
  }

  return children;
}

function AppContent() {
  const { user } = useAuth();

  return (
    <div className="min-h-screen flex flex-col bg-gray-50">
      <Header />
      <main className="flex-1">
        <Routes>
          {/* Public / Auth Routes */}
          <Route path="/role-select" element={<RoleSelect />} />
          <Route path="/login" element={<Login />} />
          <Route path="/register" element={<Register />} />

          {/* Client Routes */}
          <Route
            path="/client/dashboard"
            element={
              <ProtectedRoute role="client">
                <ClientDashboard />
              </ProtectedRoute>
            }
          />
          <Route
            path="/client/orders"
            element={
              <ProtectedRoute role="client">
                <ClientOrders />
              </ProtectedRoute>
            }
          />
          <Route
            path="/client/orders/new"
            element={
              <ProtectedRoute role="client">
                <NewOrder />
              </ProtectedRoute>
            }
          />
          <Route
            path="/client/orders/:id"
            element={
              <ProtectedRoute role="client">
                <ClientTrack />
              </ProtectedRoute>
            }
          />
          <Route
            path="/client/track"
            element={
              <ProtectedRoute role="client">
                <ClientTrack />
              </ProtectedRoute>
            }
          />
          <Route
            path="/client/shop"
            element={
              <ProtectedRoute role="client">
                <NewOrder />
              </ProtectedRoute>
            }
          />
          <Route
            path="/client/profile"
            element={
              <ProtectedRoute role="client">
                <ProfileSetup />
              </ProtectedRoute>
            }
          />
          <Route
            path="/client/profile-setup"
            element={
              <ProtectedRoute role="client">
                <ProfileSetup />
              </ProtectedRoute>
            }
          />

          {/* Fallback Redirect */}
          <Route
            path="/"
            element={
              user ? (
                <Navigate to={user.role === 'rider' ? '/rider/dashboard' : '/client/dashboard'} replace />
              ) : (
                <Navigate to="/role-select" replace />
              )
            }
          />
          <Route
            path="*"
            element={
              user ? (
                <Navigate to={user.role === 'rider' ? '/rider/dashboard' : '/client/dashboard'} replace />
              ) : (
                <Navigate to="/role-select" replace />
              )
            }
          />
        </Routes>
      </main>
      <BottomNavigation />
    </div>
  );
}

export default function App() {
  return (
    <Router>
      <AuthProvider>
        <AppContent />
      </AuthProvider>
    </Router>
  );
}
