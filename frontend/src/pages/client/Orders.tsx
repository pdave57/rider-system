import { useState } from 'react';
import { Package, ShoppingBag, Plus, Filter, X, MapPin, Clock, DollarSign, ChevronRight } from 'lucide-react';
import { useOrders } from '../../hooks/useOrders';
import { useAuth } from '../../contexts/AuthContext';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';

const statusConfig = {
  pending: { label: 'Pending', color: 'badge-warning' },
  'searching-rider': { label: 'Finding Rider', color: 'bg-blue-100 text-blue-700' },
  'rider-assigned': { label: 'Rider Assigned', color: 'bg-purple-100 text-purple-700' },
  'rider-en-route-pickup': { label: 'En Route to Pickup', color: 'bg-orange-100 text-orange-700' },
  'at-pickup': { label: 'At Pickup', color: 'bg-cyan-100 text-cyan-700' },
  'rider-en-route-dropoff': { label: 'En Route to Dropoff', color: 'bg-indigo-100 text-indigo-700' },
  delivered: { label: 'Delivered', color: 'badge-success' },
  cancelled: { label: 'Cancelled', color: 'badge-danger' },
};

const typeConfig = {
  dispatch: { label: 'Dispatch', icon: Package, color: 'bg-sky-blue-100 text-sky-blue-600' },
  'shop-for-me': { label: 'Shop for Me', icon: ShoppingBag, color: 'bg-purple-100 text-purple-600' },
};

export function ClientOrders() {
  const { user } = useAuth();
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const [filterStatus, setFilterStatus] = useState<string>('all');
  const [filterType, setFilterType] = useState<string>('all');
  const { orders, loading, fetchOrders } = useOrders(user?.id || null, 'client');

  const type = searchParams.get('type') || 'dispatch';

  const filteredOrders = orders.filter(order => {
    const statusMatch = filterStatus === 'all' || order.status === filterStatus;
    const typeMatch = filterType === 'all' || order.type === filterType;
    return statusMatch && typeMatch;
  });

  const handleNewOrder = (orderType: 'dispatch' | 'shop-for-me') => {
    navigate(`/client/orders/new?type=${orderType}`);
  };

  return (
    <div className="min-h-screen bg-gray-50 pb-24">
      <div className="bg-white border-b border-gray-200">
        <div className="container py-4">
          <div className="flex items-center justify-between mb-4">
            <h1 className="text-2xl font-bold text-navy-blue">My Orders</h1>
            <button
              onClick={() => handleNewOrder('dispatch')}
              className="btn btn-primary btn-sm"
            >
              <Plus className="icon-sm" />
              New Order
            </button>
          </div>

          <div className="flex gap-2 overflow-x-auto pb-2">
            {[
              { value: 'all', label: 'All' },
              { value: 'dispatch', label: 'Dispatch' },
              { value: 'shop-for-me', label: 'Shop for Me' },
            ].map((tab) => (
              <button
                key={tab.value}
                onClick={() => setFilterType(tab.value)}
                className={`whitespace-nowrap px-4 py-2 rounded-full text-sm font-medium transition ${
                  filterType === tab.value
                    ? 'bg-sky-blue-600 text-white'
                    : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                }`}
              >
                {tab.label}
              </button>
            ))}
            <div className="flex-1" />
            <select
              value={filterStatus}
              onChange={(e) => setFilterStatus(e.target.value)}
              className="px-3 py-2 rounded-lg border border-gray-200 text-sm bg-white"
            >
              <option value="all">All Status</option>
              <option value="pending">Pending</option>
              <option value="searching-rider">Finding Rider</option>
              <option value="rider-assigned">Rider Assigned</option>
              <option value="rider-en-route-pickup">En Route Pickup</option>
              <option value="at-pickup">At Pickup</option>
              <option value="rider-en-route-dropoff">En Route Dropoff</option>
              <option value="delivered">Delivered</option>
              <option value="cancelled">Cancelled</option>
            </select>
          </div>
        </div>
      </div>

      <div className="container px-4 py-4">
        {loading ? (
          <div className="space-y-3">
            {[1, 2, 3, 4].map(i => (
              <div key={i} className="card p-4 animate-pulse">
                <div className="h-5 bg-gray-200 rounded w-3/4 mb-3" />
                <div className="h-4 bg-gray-200 rounded w-1/2" />
              </div>
            ))}
          </div>
        ) : filteredOrders.length > 0 ? (
          <div className="space-y-3">
            {filteredOrders.map((order) => {
              const TypeIcon = typeConfig[order.type].icon;
              return (
                <Link key={order.id} to={`/client/orders/${order.id}`} className="block">
                  <div className="card p-4 hover:shadow-md transition">
                    <div className="flex items-start justify-between gap-4">
                      <div className="flex items-center gap-3 flex-1 min-w-0">
                        <div className={`w-12 h-12 rounded-xl flex items-center justify-center flex-shrink-0 ${typeConfig[order.type].color}`}>
                          <TypeIcon className="icon" />
                        </div>
                        <div className="min-w-0">
                          <div className="flex items-center gap-2">
                            <p className="font-medium text-gray-900 truncate">
                              {order.type === 'dispatch' ? 'Dispatch Order' : 'Shop for Me'}
                            </p>
                            <span className={`badge px-2 py-1 text-xs ${statusConfig[order.status].color}`}>
                              {statusConfig[order.status].label}
                            </span>
                          </div>
                          <p className="text-sm text-gray-500">{order.id} • ${order.price.toFixed(2)}</p>
                          <div className="mt-1 flex items-center gap-3 text-xs text-gray-400">
                            <span className="flex items-center gap-1">
                              <MapPin className="icon-sm" />
                              {order.pickupLocation.address?.split(',')[0] || 'Loading...'}
                            </span>
                            <span className="flex items-center gap-1">
                              <Clock className="icon-sm" />
                              {order.estimatedTime} min
                            </span>
                          </div>
                        </div>
                      </div>
                      <ChevronRight className="icon text-gray-400 flex-shrink-0" />
                    </div>
                  </div>
                </Link>
              );
            })}
          </div>


        ) : (
          <div className="card p-8 text-center">
            <Package className="icon-xl text-gray-300 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">No orders found</h3>
            <p className="text-gray-500 mb-4">
              {filterStatus !== 'all' || filterType !== 'all'
                ? 'Try adjusting your filters'
                : 'You haven\'t placed any orders yet'}
            </p>
            {(filterStatus !== 'all' || filterType !== 'all') && (
              <button
                onClick={() => { setFilterStatus('all'); setFilterType('all'); }}
                className="btn btn-outline"
              >
                <X className="icon-sm" />
                Clear Filters
              </button>
            )}
            {(filterStatus === 'all' && filterType === 'all') && (
              <div className="flex gap-3 justify-center">
                <button onClick={() => handleNewOrder('dispatch')} className="btn btn-primary">
                  <Package className="icon-sm" />
                  Order Dispatch
                </button>
                <button onClick={() => handleNewOrder('shop-for-me')} className="btn btn-secondary">
                  <ShoppingBag className="icon-sm" />
                  Shop for Me
                </button>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}