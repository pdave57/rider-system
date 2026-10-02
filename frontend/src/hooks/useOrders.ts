import { useState, useEffect, useCallback } from 'react';
import { Order, OrderStatus, Location } from '../types';
import { api } from '../services/api';

export function useOrders(userId: string | null, role: 'client' | 'rider') {
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchOrders = useCallback(async () => {
    if (!userId) return;
    setLoading(true);
    setError(null);
    try {
      const data = await api.getOrders(role, userId);
      setOrders(data);
    } catch (err) {
      setError('Failed to load orders');
    } finally {
      setLoading(false);
    }
  }, [userId, role]);

  useEffect(() => {
    fetchOrders();
    const interval = setInterval(fetchOrders, 30000);
    return () => clearInterval(interval);
  }, [fetchOrders]);

  const createOrder = async (orderData: Partial<Order>): Promise<Order> => {
    const newOrder = await api.createOrder(orderData);
    setOrders(prev => [newOrder, ...prev]);
    return newOrder;
  };

  const updateOrderStatus = async (orderId: string, status: OrderStatus): Promise<void> => {
    await api.updateOrderStatus(orderId, status);
    setOrders(prev => prev.map(o => o.id === orderId ? { ...o, status } : o));
  };

  const acceptOrder = async (orderId: string, riderId: string): Promise<void> => {
    await api.respondDispatch(orderId, 'accept');
    setOrders(prev => prev.map(o => o.id === orderId ? { ...o, riderId, status: 'rider-assigned', acceptedAt: new Date().toISOString() } : o));
  };

  const rejectOrder = async (orderId: string, riderId: string): Promise<void> => {
    await api.respondDispatch(orderId, 'reject');
    setOrders(prev => prev.map(o => o.id === orderId ? { ...o, status: 'pending', riderId: undefined } : o));
  };

  return {
    orders,
    loading,
    error,
    fetchOrders,
    createOrder,
    updateOrderStatus,
    acceptOrder,
    rejectOrder,
  };
}

export function useOrderTracking(orderId: string | null) {
  const [trackingData, setTrackingData] = useState<{
    riderLocation: Location;
    estimatedArrival: number;
    status: OrderStatus;
    route: Location[];
  } | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!orderId) return;
    
    const fetchTracking = async () => {
      setLoading(true);
      try {
        const data = await api.getTrackingData(orderId);
        if (data) setTrackingData(data);
      } catch {
        // Ignore errors for tracking
      } finally {
        setLoading(false);
      }
    };

    fetchTracking();
    const interval = setInterval(fetchTracking, 5000);
    
    return () => clearInterval(interval);

  }, [orderId]);

  return { trackingData, loading };
}

export function useNearbyRiders(location: Location | null, radiusKm: number = 1) {
  const [riders, setRiders] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!location) return;
    
    const fetchRiders = async () => {
      setLoading(true);
      try {
        const data = await api.getNearbyRiders(location, radiusKm);
        setRiders(data);
      } catch {
        setRiders([]);
      } finally {
        setLoading(false);
      }
    };

    fetchRiders();
    const interval = setInterval(fetchRiders, 10000);
    return () => clearInterval(interval);
  }, [location, radiusKm]);

  return { riders, loading };
}

export function useRiderAvailability(riderId: string | null) {
  const [isOnline, setIsOnline] = useState(false);
  const [loading, setLoading] = useState(false);

  const toggleAvailability = async () => {
    if (!riderId) return;
    setLoading(true);
    try {
      const rider = await api.setRiderAvailability(riderId, !isOnline);
      setIsOnline(rider.isOnline || false);
    } catch {
      // Error handled silently
    } finally {
      setLoading(false);
    }
  };

  return { isOnline, loading, toggleAvailability };
}

export function useRiderLocation(riderId: string | null) {
  const updateLocation = useCallback(async (location: Location) => {
    if (!riderId) return;
    await api.updateRiderLocation(riderId, location);
  }, [riderId]);

  return { updateLocation };
}