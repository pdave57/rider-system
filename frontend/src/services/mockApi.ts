import { Order, OrderStatus, Location, User, UserRole } from '../types';

const mockOrders: Order[] = [
  {
    id: 'ORD-001',
    clientId: '1',
    type: 'dispatch',
    status: 'pending',
    pickupLocation: { lat: 40.7128, lng: -74.0060, address: '123 Main St, NYC' },
    dropoffLocation: { lat: 40.7589, lng: -73.9851, address: '456 Broadway, NYC' },
    items: [
      { id: '1', name: 'Documents', quantity: 1, price: 0 },
    ],
    distance: 2.5,
    estimatedTime: 15,
    price: 12.50,
    paymentStatus: 'paid',
    paymentMethod: 'card',
    createdAt: new Date(Date.now() - 300000).toISOString(),
  },
  {
    id: 'ORD-002',
    clientId: '1',
    type: 'shop-for-me',
    status: 'delivered',
    pickupLocation: { lat: 40.7505, lng: -73.9934, address: 'Whole Foods, NYC' },
    dropoffLocation: { lat: 40.7128, lng: -74.0060, address: '123 Main St, NYC' },
    items: [
      { id: '1', name: 'Organic Milk', quantity: 2, price: 4.99 },
      { id: '2', name: 'Whole Wheat Bread', quantity: 1, price: 3.49 },
      { id: '3', name: 'Free-range Eggs', quantity: 1, price: 5.99 },
    ],
    distance: 1.8,
    estimatedTime: 20,
    price: 24.95,
    paymentStatus: 'paid',
    paymentMethod: 'wallet',
    notes: 'Leave at door if not home',
    createdAt: new Date(Date.now() - 86400000).toISOString(),
    acceptedAt: new Date(Date.now() - 86000000).toISOString(),
    pickedUpAt: new Date(Date.now() - 84000000).toISOString(),
    deliveredAt: new Date(Date.now() - 82000000).toISOString(),
  },
];

const mockRiders: User[] = [
  {
    id: 'rider-1',
    email: 'rider1@example.com',
    firstName: 'Mike',
    lastName: 'Johnson',
    phone: '+1234567891',
    role: 'rider',
    isVerified: true,
    profileComplete: true,
    createdAt: new Date().toISOString(),
    licenseNumber: 'DL987654321',
    vehicleType: 'Motorcycle',
    vehiclePlate: 'XYZ-7890',
    rating: 4.9,
    totalDeliveries: 234,
    isOnline: true,
    currentLocation: { lat: 40.7140, lng: -74.0070 },
  },
  {
    id: 'rider-2',
    email: 'rider2@example.com',
    firstName: 'Sarah',
    lastName: 'Williams',
    phone: '+1234567892',
    role: 'rider',
    isVerified: true,
    profileComplete: true,
    createdAt: new Date().toISOString(),
    licenseNumber: 'DL111222333',
    vehicleType: 'Bicycle',
    vehiclePlate: 'N/A',
    rating: 4.7,
    totalDeliveries: 189,
    isOnline: true,
    currentLocation: { lat: 40.7100, lng: -74.0090 },
  },
  {
    id: 'rider-3',
    email: 'rider3@example.com',
    firstName: 'David',
    lastName: 'Brown',
    phone: '+1234567893',
    role: 'rider',
    isVerified: true,
    profileComplete: true,
    createdAt: new Date().toISOString(),
    licenseNumber: 'DL444555666',
    vehicleType: 'Scooter',
    vehiclePlate: 'DEF-5678',
    rating: 4.8,
    totalDeliveries: 156,
    isOnline: false,
    currentLocation: { lat: 40.7200, lng: -74.0150 },
  },
];

const calculateDistance = (loc1: Location, loc2: Location): number => {
  const R = 6371;
  const dLat = (loc2.lat - loc1.lat) * Math.PI / 180;
  const dLng = (loc2.lng - loc1.lng) * Math.PI / 180;
  const a = Math.sin(dLat/2) * Math.sin(dLat/2) +
    Math.cos(loc1.lat * Math.PI / 180) * Math.cos(loc2.lat * Math.PI / 180) *
    Math.sin(dLng/2) * Math.sin(dLng/2);
  const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1-a));
  return R * c;
};

export const mockApi = {
  async getOrders(role: 'client' | 'rider', userId: string): Promise<Order[]> {
    await new Promise(r => setTimeout(r, 500));
    if (role === 'client') {
      return mockOrders.filter(o => o.clientId === userId);
    }
    return mockOrders.filter(o => o.riderId === userId || o.status === 'pending');
  },

  async createOrder(orderData: Partial<Order>): Promise<Order> {
    await new Promise(r => setTimeout(r, 800));
    const newOrder: Order = {
      id: `ORD-${Date.now()}`,
      clientId: orderData.clientId || '1',
      type: orderData.type || 'dispatch',
      status: 'pending',
      pickupLocation: orderData.pickupLocation!,
      dropoffLocation: orderData.dropoffLocation!,
      items: orderData.items || [],
      distance: orderData.distance || 0,
      estimatedTime: orderData.estimatedTime || 0,
      price: orderData.price || 0,
      paymentStatus: 'pending',
      paymentMethod: orderData.paymentMethod || 'cash',
      notes: orderData.notes,
      createdAt: new Date().toISOString(),
    };
    mockOrders.unshift(newOrder);
    
    this.findAndAssignRider(newOrder);
    return newOrder;
  },

  async findAndAssignRider(order: Order): Promise<void> {
    const nearbyRiders = mockRiders.filter(r => 
      r.isOnline && 
      r.isVerified && 
      calculateDistance(r.currentLocation!, order.pickupLocation) <= 1
    );
    
    if (nearbyRiders.length > 0) {
      const assignedRider = nearbyRiders[0];
      order.riderId = assignedRider.id;
      order.status = 'rider-assigned';
      
      setTimeout(() => {
        order.status = 'rider-en-route-pickup';
      }, 5000);
      
      setTimeout(() => {
        order.status = 'at-pickup';
      }, 15000);
      
      setTimeout(() => {
        order.status = 'rider-en-route-dropoff';
      }, 20000);
      
      setTimeout(() => {
        order.status = 'delivered';
        order.deliveredAt = new Date().toISOString();
        assignedRider.totalDeliveries = (assignedRider.totalDeliveries || 0) + 1;
      }, 35000);
    } else {
      order.status = 'searching-rider';
      setTimeout(() => this.findAndAssignRider(order), 10000);
    }
  },

  async getNearbyRiders(location: Location, radiusKm: number = 1): Promise<User[]> {
    await new Promise(r => setTimeout(r, 500));
    return mockRiders.filter(r => 
      r.isOnline && 
      r.isVerified && 
      calculateDistance(r.currentLocation!, location) <= radiusKm
    );
  },

  async acceptOrder(orderId: string, riderId: string): Promise<Order> {
    await new Promise(r => setTimeout(r, 500));
    const order = mockOrders.find(o => o.id === orderId);
    if (order) {
      order.riderId = riderId;
      order.status = 'rider-assigned';
      order.acceptedAt = new Date().toISOString();
    }
    return order!;
  },

  async rejectOrder(orderId: string, riderId: string): Promise<Order> {
    await new Promise(r => setTimeout(r, 500));
    const order = mockOrders.find(o => o.id === orderId);
    if (order) {
      order.status = 'pending';
      order.riderId = undefined;
    }
    return order!;
  },

  async updateOrderStatus(orderId: string, status: OrderStatus): Promise<Order> {
    await new Promise(r => setTimeout(r, 300));
    const order = mockOrders.find(o => o.id === orderId);
    if (order) {
      order.status = status;
      if (status === 'delivered') order.deliveredAt = new Date().toISOString();
      if (status === 'cancelled') order.cancelledAt = new Date().toISOString();
    }
    return order!;
  },

  async getTrackingData(orderId: string) {
    await new Promise(r => setTimeout(r, 300));
    const order = mockOrders.find(o => o.id === orderId);
    if (!order || !order.riderId) return null;
    
    const rider = mockRiders.find(r => r.id === order.riderId);
    return {
      orderId,
      riderLocation: rider?.currentLocation || order.pickupLocation,
      estimatedArrival: 10,
      status: order.status,
      route: [order.pickupLocation, order.dropoffLocation],
    };
  },

  async updateRiderLocation(riderId: string, location: Location): Promise<void> {
    const rider = mockRiders.find(r => r.id === riderId);
    if (rider) rider.currentLocation = location;
  },

  async setRiderAvailability(riderId: string, isOnline: boolean): Promise<User> {
    await new Promise(r => setTimeout(r, 300));
    const rider = mockRiders.find(r => r.id === riderId);
    if (rider) rider.isOnline = isOnline;
    return rider!;
  },
};

export function useMockApi() {
  return mockApi;
}