import { Menu, X, LogOut, User, Settings, Bell } from 'lucide-react';
import { useAuth } from '../contexts/AuthContext';
import { useNavigate, useLocation } from 'react-router-dom';
import { useState, ReactNode } from 'react';

interface HeaderProps {
  title?: string;
  showBack?: boolean;
  children?: ReactNode;
}

export function Header({ title, showBack = false, children }: HeaderProps) {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [showMenu, setShowMenu] = useState(false);

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  const isAuthPage = ['/login', '/register', '/role-select'].includes(location.pathname);
  const isRiderDashboard = location.pathname.startsWith('/rider');

  if (isAuthPage) {
    return null;
  }

  return (
    <header className="bg-white border-b border-gray-200 sticky top-0 z-50">
      <div className="container">
        <div className="flex items-center justify-between h-16">
          <div className="flex items-center gap-3">
            {showBack && (
              <button
                onClick={() => navigate(-1)}
                className="p-2 rounded-lg hover:bg-gray-100 transition"
                aria-label="Go back"
              >
                <svg className="icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
                </svg>
              </button>
            )}
            <h1 className="text-xl font-bold text-navy-blue">{title || 'Rider System'}</h1>
          </div>

          <div className="flex items-center gap-2">
            {children}
            
            {user && !isRiderDashboard && (
              <>
                <button className="p-2 rounded-lg hover:bg-gray-100 transition relative" aria-label="Notifications">
                  <Bell className="icon text-gray-600" />
                  <span className="absolute top-1 right-1 w-4 h-4 bg-red-500 text-white text-xs rounded-full flex items-center justify-center">3</span>
                </button>
              </>
            )}

            {user && (
              <div className="relative">
                <button
                  onClick={() => setShowMenu(!showMenu)}
                  className="flex items-center gap-2 p-1 rounded-lg hover:bg-gray-100 transition"
                  aria-label="User menu"
                >
                  {user.avatar ? (
                    <img src={user.avatar} alt="" className="avatar" />
                  ) : (
                    <div className="avatar-placeholder">
                      {user.firstName[0]}{user.lastName[0]}
                    </div>
                  )}
                  <span className="hidden sm:block font-medium text-gray-700">
                    {user.firstName}
                  </span>
                  <svg className="icon text-gray-500 hidden sm:block" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                  </svg>
                </button>

                {showMenu && (
                  <div className="absolute right-0 top-full mt-2 w-48 bg-white rounded-lg shadow-lg border border-gray-200 py-1 z-50">
                    <div className="px-4 py-2 border-b border-gray-100">
                      <p className="font-medium text-gray-900">{user.firstName} {user.lastName}</p>
                      <p className="text-sm text-gray-500 capitalize">{user.role}</p>
                    </div>
                    <button
                      onClick={() => navigate(user.role === 'rider' ? '/rider/profile' : '/client/profile')}
                      className="w-full px-4 py-2 text-left hover:bg-gray-50 flex items-center gap-2"
                    >
                      <User className="icon text-gray-500" />
                      Profile
                    </button>
                    <button
                      onClick={() => navigate('/settings')}
                      className="w-full px-4 py-2 text-left hover:bg-gray-50 flex items-center gap-2"
                    >
                      <Settings className="icon text-gray-500" />
                      Settings
                    </button>
                    <hr className="my-1 border-gray-100" />
                    <button
                      onClick={handleLogout}
                      className="w-full px-4 py-2 text-left hover:bg-gray-50 flex items-center gap-2 text-red-600"
                    >
                      <LogOut className="icon" />
                      Logout
                    </button>
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
      </div>
      
      {showMenu && (
        <div 
          className="fixed inset-0 z-40" 
          onClick={() => setShowMenu(false)}
          aria-hidden="true"
        />
      )}
    </header>
  );
}