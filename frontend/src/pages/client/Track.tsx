import { useState, useEffect } from 'react';
import { MapPin, Bike, Package, ShoppingBag, Clock, CheckCircle, XCircle, AlertCircle, Navigation, RefreshCw } from 'lucide-react';
import { useAuth } from '../../contexts/AuthContext';
import { useOrders, useOrderTracking } from '../../hooks/useOrders';

const statusFlow = [
  { key: 'pending', label: 'Order Placed', icon: Package },
  { key: 'searching-rider', label: 'Finding Rider', icon: AlertCircle },
  { key: 'rider-assigned', label: 'Rider Assigned', icon: Bike },
  { key: 'rider-en-route-pickup', label: 'En Route to Pickup', icon: Navigation },
  { key: 'at-pickup', label: 'At Pickup', icon: MapPin },
  { key: 'rider-en-route-dropoff', label: 'En Route to Dropoff', icon: Navigation },
  { key: 'delivered', label: 'Delivered', icon: CheckCircle },
];

const statusConfig = {
  pending: { label: 'Pending', color: 'text-yellow-600 bg-yellow-100' },
  'searching-rider': { label: 'Finding Rider', color: 'text-blue-600 bg-blue-100' },
  'rider-assigned': { label: 'Rider Assigned', color: 'text-purple-600 bg-purple-100' },
  'rider-en-route-pickup': { label: 'En Route to Pickup', color: 'text-orange-600 bg-orange-100' },
  'at-pickup': { label: 'At Pickup', color: 'text-cyan-600 bg-cyan-100' },
  'rider-en-route-dropoff': { label: 'En Route to Dropoff', color: 'text-indigo-600 bg-indigo-100' },
  delivered: { label: 'Delivered', color: 'text-green-600 bg-green-100' },
  cancelled: { label: 'Cancelled', color: 'text-red-600 bg-red-100' },
};

export function ClientTrack() {
  const { user } = useAuth();
  const { orders, loading } = useOrders(user?.id || null, 'client');
  const [selectedOrderId, setSelectedOrderId] = useState<string | null>(null);
  const { trackingData, loading: trackingLoading } = useOrderTracking(selectedOrderId);

  const activeOrders = orders.filter(o => !['delivered', 'cancelled'].includes(o.status));
  const selectedOrder = orders.find(o => o.id === selectedOrderId) || activeOrders[0];

  useEffect(() => {
    if (activeOrders.length > 0 && !selectedOrderId) {
      setSelectedOrderId(activeOrders[0].id);
    }
  }, [activeOrders, selectedOrderId]);

  if (!selectedOrder) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
        <div className="text-center">
          <MapPin className="icon-xl text-gray-300 mx-auto mb-4" />
          <h2 className="text-xl font-semibold text-gray-900 mb-2">No active orders</h2>
          <p className="text-gray-500 mb-6">You don't have any orders currently being tracked</p>
          <a href="/client/orders/new?type=dispatch" className="btn btn-primary">Create New Order</a>
        </div>
      </div>
    );
  }

  const currentStatusIndex = statusFlow.findIndex(s => s.key === selectedOrder.status);
  const isDelivered = selectedOrder.status === 'delivered';
  const isCancelled = selectedOrder.status === 'cancelled';

  return (
    <div className="min-h-screen bg-gray-50 pb-24">
      <div className="bg-white border-b border-gray-200">
        <div className="container py-4">
          <h1 className="text-2xl font-bold text-navy-blue mb-4">Track Order</h1>
          
          {activeOrders.length > 1 && (
            <div className="mb-4">
              <select
                value={selectedOrderId || ''}
                onChange={(e) => setSelectedOrderId(e.target.value)}
                className="w-full px-4 py-2 rounded-lg border border-gray-200 bg-white text-sm"
              >
                {activeOrders.map(order => (
                  <option key={order.id} value={order.id}>
                    {order.id} - {order.type === 'dispatch' ? 'Dispatch' : 'Shop for Me'} - ${order.price.toFixed(2)}
                  </option>
                ))}
              </select>
            </div>
          )}

          <div className="card">
            <div className="card-body p-4">
              <div className="flex items-start justify-between gap-4 mb-4">
                <div className="flex items-center gap-3">
                  <div className={`w-12 h-12 rounded-xl flex items-center justify-center ${statusConfig[selectedOrder.status].color}`}>
                    {selectedOrder.type === 'dispatch' ? <Package className="icon" /> : <ShoppingBag className="icon" />}
                  </div>
                  <div>
                    <p className="font-semibold text-gray-900">{selectedOrder.id}</p>
                    <p className="text-sm text-gray-500">{selectedOrder.type === 'dispatch' ? 'Dispatch Order' : 'Shop for Me'}</p>
                  </div>
                </div>
                <span className={`badge px-3 py-1 ${statusConfig[selectedOrder.status].color}`}>
                  {statusConfig[selectedOrder.status].label}
                </span>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="p-3 bg-gray-50 rounded-lg">
                  <p className="text-xs text-gray-500">Pickup</p>
                  <p className="font-medium text-gray-900 truncate">{selectedOrder.pickupLocation.address}</p>
                </div>
                <div className="p-3 bg-gray-50 rounded-lg">
                  <p className="text-xs text-gray-500">Dropoff</p>
                  <p className="font-medium text-gray-900 truncate">{selectedOrder.dropoffLocation.address}</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div className="container px-4 py-4">
        <div className="card mb-6">
          <div className="card-body p-4">
            <div className="relative">
              <div className="absolute left-5 top-0 bottom-0 w-0.5 bg-gray-200" />
              <div className="absolute left-5 top-0 h-4 w-0.5 bg-sky-blue-600" />
              
              {statusFlow.map((status, index) => {
                const isCompleted = index < currentStatusIndex;
                const isCurrent = index === currentStatusIndex && !isDelivered && !isCancelled;
                const Icon = status.icon;
                
                return (
                  <div key={status.key} className="relative flex items-start gap-3 mb-6 last:mb-0">
                    <div className={`relative flex-shrink-0 w-10 h-10 rounded-full flex items-center justify-center z-10 transition-all ${
                      isCompleted ? 'bg-sky-blue-600 text-white' :
                      isCurrent ? 'bg-sky-blue-600 text-white ring-4 ring-sky-blue-200 animate-pulse' :
                      'bg-gray-200 text-gray-400'
                    }`}>
                      {isCompleted ? <CheckCircle className="icon" /> : <Icon className="icon" />}
                    </div>
                    <div className="flex-1 pt-1">
                      <p className={`font-medium ${isCurrent ? 'text-navy-blue' : isCompleted ? 'text-gray-900' : 'text-gray-500'}`}>
                        {status.label}
                      </p>
                      {isCurrent && (
                        <p className="text-sm text-sky-blue-600 mt-0.5">
                          {trackingData ? `Rider arriving in ~${trackingData.estimatedArrival} min` : 'Updating...'}
                        </p>
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        </div>

        {trackingData && !isDelivered && !isCancelled && (
          <div className="card mb-6">
            <div className="card-header">
              <h3 className="font-semibold text-navy-blue">Live Tracking</h3>
            </div>
            <div className="card-body p-4">
              <div className="aspect-video bg-gray-100 rounded-lg relative overflow-hidden">
                <div className="absolute inset-0 flex items-center justify-center text-gray-400">
                  <div className="text-center">
                    <MapPin className="icon-xl mx-auto mb-2" />
                    <p className="text-sm">Map View</p>
                    <p className="text-xs">Rider: {trackingData.riderLocation.lat.toFixed(4)}, {trackingData.riderLocation.lng.toFixed(4)}</p>
                  </div>
                </div>
                
                <div className="absolute bottom-4 left-4 right-4 bg-white/90 backdrop-blur rounded-lg p-3 shadow-lg">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <div className="w-10 h-10 rounded-full bg-sky-blue-100 flex items-center justify-center">
                        <Bike className="icon text-sky-blue-600" />
                      </div>
                      <div>
                        <p className="font-medium text-gray-900">Your Rider</p>
                        <p className="text-sm text-gray-500">On the way</p>
                      </div>
                    </div>
                    <div className="text-right">
                      <p className="text-2xl font-bold text-sky-blue-600">{trackingData.estimatedArrival} min</p>
                      <p className="text-xs text-gray-500">Est. Arrival</p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}

        {(isDelivered || isCancelled) && (
          <div className="card">
            <div className="card-body p-6 text-center">
              <div className={`w-16 h-16 rounded-full flex items-center justify-center mx-auto mb-4 ${isDelivered ? 'bg-green-100 text-green-600' : 'bg-red-100 text-red-600'}`}>
                {isDelivered ? <CheckCircle className="icon-xl" /> : <XCircle className="icon-xl" />}
              </div>
              <h3 className="text-xl font-bold text-navy-blue mb-2">
                {isDelivered ? 'Order Delivered!' : 'Order Cancelled'}
              </h3>
              <p className="text-gray-500 mb-6">
                {isDelivered 
                  ? 'Your order has been successfully delivered. Thank you for using our service!'
                  : 'This order was cancelled. Contact support if you have questions.'}
              </p>
              <a href="/client/orders" className="btn btn-primary">View All Orders</a>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}