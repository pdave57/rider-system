export type UserRole = 'client' | 'rider';

export interface Location {
  lat: number;
  lng: number;
  address?: string;
}

export interface Order {
  id: string;
  clientId: string;
  riderId?: string;
  type: 'dispatch' | 'shop-for-me';
  status: OrderStatus;
  pickupLocation: Location;
  dropoffLocation: Location;
  items: OrderItem[];
  distance: number;
  estimatedTime: number;
  price: number;
  paymentStatus: 'pending' | 'paid' | 'failed';
  paymentMethod: 'cash' | 'card' | 'wallet';
  notes?: string;
  createdAt: string;
  acceptedAt?: string;
  pickedUpAt?: string;
  deliveredAt?: string;
  cancelledAt?: string;
  cancellationReason?: string;
}

export type OrderStatus = 
  | 'pending'
  | 'searching-rider'
  | 'rider-assigned'
  | 'rider-en-route-pickup'
  | 'at-pickup'
  | 'rider-en-route-dropoff'
  | 'delivered'
  | 'cancelled';

export interface OrderItem {
  id: string;
  name: string;
  quantity: number;
  price: number;
  image?: string;
  notes?: string;
}

export interface User {
  id: string;
  email: string;
  firstName: string;
  lastName: string;
  phone: string;
  role: UserRole;
  avatar?: string;
  isVerified: boolean;
  profileComplete: boolean;
  createdAt: string;
  
  // Client specific
  address?: string;
  preferences?: {
    notifications: boolean;
    smsAlerts: boolean;
    emailAlerts: boolean;
  };
  
  // Rider specific
  licenseNumber?: string;
  licenseExpiry?: string;
  vehicleType?: string;
  vehiclePlate?: string;
  rating?: number;
  totalDeliveries?: number;
  isOnline?: boolean;
  currentLocation?: Location;
}

export interface AuthState {
  user: User | null;
  loading: boolean;
  error: string | null;
}

export interface RegisterData {
  email: string;
  password: string;
  confirmPassword: string;
  firstName: string;
  lastName: string;
  phone: string;
  role: UserRole;
}

export interface RiderVerificationData {
  licenseNumber: string;
  licenseExpiry: string;
  vehicleType: string;
  vehiclePlate: string;
  idDocument: File;
  selfie: File;
}

export interface TrackingData {
  orderId: string;
  riderLocation: Location;
  estimatedArrival: number;
  status: OrderStatus;
  route: Location[];
}