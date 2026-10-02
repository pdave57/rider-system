import { Home, Package, ShoppingBag, MapPin, User, Menu, Grid, LogOut, Settings } from 'lucide-react';
import { useAuth } from '../contexts/AuthContext';
import { useNavigate, useLocation } from 'react-router-dom';
import { useState } from 'react';

const clientNavItems = [
  { path: '/client/dashboard', label: 'Home', icon: Home },
  { path: '/client/orders', label: 'Orders', icon: Package },
  { path: '/client/shop', label: 'Shop', icon: ShoppingBag },
  { path: '/client/track', label: 'Track', icon: MapPin },
  { path: '/client/profile', label: 'Profile', icon: User },
];

const riderNavItems = [
  { path: '/rider/dashboard', label: 'Dashboard', icon: Home },
  { path: '/rider/orders', label: 'Orders', icon: Package },
  { path: '/rider/earnings', label: 'Earnings', icon: Menu },
  { path: '/rider/profile', label: 'Profile', icon: User },
];

export function BottomNavigation() {
  const { user } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [showMore, setShowMore] = useState(false);

  if (!user) return null;

  const navItems = user.role === 'rider' ? riderNavItems : clientNavItems;
  const currentPath = location.pathname;

  return (
    <>
      <nav className="fixed bottom-0 left-0 right-0 bg-white border-t border-gray-200 z-50 md:hidden">
        <div className="grid grid-cols-5 gap-1 px-2 py-2">
          {navItems.map((item, index) => {
            const isActive = currentPath === item.path || (item.path !== '/client/dashboard' && currentPath.startsWith(item.path));
            const Icon = item.icon;
            return (
              <button
                key={item.path}
                onClick={() => navigate(item.path)}
                className={`flex flex-col items-center gap-1 py-2 px-2 rounded-lg transition ${
                  isActive
                    ? 'bg-sky-blue-50 text-sky-blue-600'
                    : 'text-gray-500 hover:bg-gray-50 hover:text-gray-700'
                }`}
                aria-current={isActive ? 'page' : undefined}
              >
                <Icon className={`icon ${isActive ? 'fill-current' : ''}`} />
                <span className="text-xs font-medium">{item.label}</span>
              </button>
            );
          })}
        </div>
      </nav>

      <div className="fixed bottom-0 left-0 right-0 bg-white border-t border-gray-200 z-50 md:hidden pb-16">
        <div className="grid grid-cols-4 gap-1 px-2 py-2">
          {navItems.slice(0, 4).map((item) => {
            const isActive = currentPath === item.path || (item.path !== '/client/dashboard' && currentPath.startsWith(item.path));
            const Icon = item.icon;
            return (
              <button
                key={item.path}
                onClick={() => navigate(item.path)}
                className={`flex flex-col items-center gap-1 py-2 px-2 rounded-lg transition ${
                  isActive
                    ? 'bg-sky-blue-50 text-sky-blue-600'
                    : 'text-gray-500 hover:bg-gray-50 hover:text-gray-700'
                }`}
              >
                <Icon className={`icon ${isActive ? 'fill-current' : ''}`} />
                <span className="text-xs font-medium">{item.label}</span>
              </button>
            );
          })}
          <button
            onClick={() => setShowMore(!showMore)}
            className="flex flex-col items-center gap-1 py-2 px-2 rounded-lg transition text-gray-500 hover:bg-gray-50"
          >
            <Grid className="icon" />
            <span className="text-xs font-medium">More</span>
          </button>
        </div>

        {showMore && (
          <div className="bg-white border-t border-gray-200 px-2 py-2">
            <div className="grid grid-cols-2 gap-2">
              {navItems.slice(4).map((item) => {
                const isActive = currentPath === item.path;
                const Icon = item.icon;
                return (
                  <button
                    key={item.path}
                    onClick={() => { navigate(item.path); setShowMore(false); }}
                    className={`flex flex-col items-center gap-1 py-3 px-2 rounded-lg transition ${
                      isActive
                        ? 'bg-sky-blue-50 text-sky-blue-600'
                        : 'text-gray-500 hover:bg-gray-50 hover:text-gray-700'
                    }`}
                  >
                    <Icon className={`icon-lg ${isActive ? 'fill-current' : ''}`} />
                    <span className="text-sm font-medium">{item.label}</span>
                  </button>
                );
              })}
              <button
                onClick={() => { navigate('/settings'); setShowMore(false); }}
                className="flex flex-col items-center gap-1 py-3 px-2 rounded-lg transition text-gray-500 hover:bg-gray-50"
              >
                <Settings className="icon-lg" />
                <span className="text-sm font-medium">Settings</span>
              </button>
              <button
                onClick={() => { navigate('/login'); setShowMore(false); }}
                className="flex flex-col items-center gap-1 py-3 px-2 rounded-lg transition text-red-600 hover:bg-red-50"
              >
                <LogOut className="icon-lg" />
                <span className="text-sm font-medium">Logout</span>
              </button>
            </div>
          </div>
        )}
      </div>
    </>
  );
}