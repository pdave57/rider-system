import { Package, ShoppingBag, MapPin, User, Plus, Clock, MapPin as MapPinIcon, DollarSign, Bell } from 'lucide-react';
import { useAuth } from '../../contexts/AuthContext';
import { useOrders } from '../../hooks/useOrders';
import { Link, useNavigate } from 'react-router-dom';

const quickActions = [
  { label: 'Order Dispatch', icon: Package, path: '/client/orders/new?type=dispatch', color: 'bg-sky-blue-100 text-sky-blue-600', iconBg: 'bg-sky-blue-100' },
  { label: 'Shop for Me', icon: ShoppingBag, path: '/client/shop', color: 'bg-purple-100 text-purple-600', iconBg: 'bg-purple-100' },
  { label: 'Track Order', icon: MapPin, path: '/client/track', color: 'bg-green-100 text-green-600', iconBg: 'bg-green-100' },
  { label: 'Profile', icon: User, path: '/client/profile', color: 'bg-gray-100 text-gray-600', iconBg: 'bg-gray-100' },
];

const statusConfig = {
  pending: { label: 'Pending', color: 'bg-yellow-100 text-yellow-700' },
  'searching-rider': { label: 'Finding Rider', color: 'bg-blue-100 text-blue-700' },
  'rider-assigned': { label: 'Rider Assigned', color: 'bg-purple-100 text-purple-700' },
  'rider-en-route-pickup': { label: 'En Route to Pickup', color: 'bg-orange-100 text-orange-700' },
  'at-pickup': { label: 'At Pickup', color: 'bg-cyan-100 text-cyan-700' },
  'rider-en-route-dropoff': { label: 'En Route to Dropoff', color: 'bg-indigo-100 text-indigo-700' },
  delivered: { label: 'Delivered', color: 'bg-green-100 text-green-700' },
  cancelled: { label: 'Cancelled', color: 'bg-red-100 text-red-700' },
};

export function ClientDashboard() {
  const { user } = useAuth();
  const navigate = useNavigate();
  const { orders, loading, fetchOrders } = useOrders(user?.id || null, 'client');

  const activeOrders = orders.filter(o => !['delivered', 'cancelled'].includes(o.status));
  const recentOrders = orders.slice(0, 3);

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="bg-white">
        <div className="container py-6">
          <div className="flex items-center justify-between mb-6">
            <div>
              <h1 className="text-2xl font-bold text-navy-blue">Welcome back, {user?.firstName}!</h1>
              <p className="text-gray-500">What would you like to do today?</p>
            </div>
            <div className="relative">
              <button className="p-2 rounded-lg hover:bg-gray-100 transition relative">
                <Bell className="icon text-gray-600" />
                <span className="absolute top-1 right-1 w-4 h-4 bg-red-500 text-white text-xs rounded-full flex items-center justify-center">3</span>
              </button>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-3 mb-6">
            {quickActions.map((action, i) => {
              const Icon = action.icon;
              return (
                <Link
                  key={action.label}
                  to={action.path}
                  className="flex flex-col items-center gap-2 p-4 rounded-xl bg-white border border-gray-200 hover:border-sky-blue-200 hover:shadow-md transition"
                >
                  <div className={`w-12 h-12 rounded-xl flex items-center justify-center ${action.iconBg}`}>
                    <Icon className="icon text-sky-blue-600" />
                  </div>
                  <span className="text-sm font-medium text-gray-700">{action.label}</span>
                </Link>
              );
            })}
          </div>
        </div>
      </div>

      <div className="container px-4 pb-24">
        {activeOrders.length > 0 && (
          <div className="mb-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold text-navy-blue">Active Orders</h2>
              <Link to="/client/orders" className="text-sm text-sky-blue-600 font-medium">View all</Link>
            </div>
            <div className="space-y-3">
              {activeOrders.map((order) => (
                <Link key={order.id} to={`/client/orders/${order.id}`} className="block">
                  <div className="card p-4">
                    <div className="flex items-start justify-between gap-4">
                      <div className="flex items-center gap-3">
                        <div className={`w-12 h-12 rounded-xl flex items-center justify-center ${statusConfig[order.status].color}`}>
                          {order.type === 'dispatch' ? <Package className="icon" /> : <ShoppingBag className="icon" />}
                        </div>
                        <div>
                          <p className="font-medium text-gray-900">{order.type === 'dispatch' ? 'Dispatch Order' : 'Shop for Me'}</p>
                          <p className="text-sm text-gray-500">{order.id} • ${order.price.toFixed(2)}</p>
                        </div>
                      </div>
                      <span className={`badge ${statusConfig[order.status].color.replace('bg-', 'badge-').replace('text-', '')}`}>
                        {statusConfig[order.status].label}
                      </span>
                    </div>
                    <div className="mt-3 flex items-center gap-4 text-sm text-gray-500">
                      <span className="flex items-center gap-1">
                        <MapPinIcon className="icon-sm" />
                        {order.pickupLocation.address?.split(',')[0]}
                      </span>
                      <span className="flex items-center gap-1">
                        <Clock className="icon-sm" />
                        {order.estimatedTime} min
                      </span>
                    </div>
                  </div>
                </Link>
              ))}
            </div>
          </div>
        )}

        <div>
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold text-navy-blue">Recent Orders</h2>
            <Link to="/client/orders" className="text-sm text-sky-blue-600 font-medium">View all</Link>
          </div>
          {loading ? (
            <div className="space-y-3">
              {[1, 2, 3].map(i => (
                <div key={i} className="card p-4 animate-pulse">
                  <div className="h-5 bg-gray-200 rounded w-3/4 mb-3" />
                  <div className="h-4 bg-gray-200 rounded w-1/2" />
                </div>
              ))}
            </div>
          ) : recentOrders.length > 0 ? (
            <div className="space-y-3">
              {recentOrders.map((order) => (
                <Link key={order.id} to={`/client/orders/${order.id}`} className="block">
                  <div className="card p-4">
                    <div className="flex items-center justify-between gap-4">
                      <div className="flex items-center gap-3">
                        <div className={`w-12 h-12 rounded-xl flex items-center justify-center ${statusConfig[order.status].color}`}>
                          {order.type === 'dispatch' ? <Package className="icon" /> : <ShoppingBag className="icon" />}
                        </div>
                        <div>
                          <p className="font-medium text-gray-900">{order.type === 'dispatch' ? 'Dispatch Order' : 'Shop for Me'}</p>
                          <p className="text-sm text-gray-500">{order.id} • ${order.price.toFixed(2)}</p>
                        </div>
                      </div>
                      <span className={`badge ${statusConfig[order.status].color.replace('bg-', 'badge-').replace('text-', '')}`}>
                        {statusConfig[order.status].label}
                      </span>
                    </div>
                  </div>
                </Link>
              ))}
            </div>
          ) : (
            <div className="card p-8 text-center">
              <Package className="icon-xl text-gray-300 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-gray-900 mb-2">No orders yet</h3>
              <p className="text-gray-500 mb-4">Start by creating your first order</p>
              <Link to="/client/orders/new?type=dispatch" className="btn btn-primary">Create Order</Link>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}