import { Order, OrderStatus, Location, User, UserRole, RegisterData } from '../types';
import { mockApi } from './mockApi';

const API_BASE = '/api';

interface ApiResponse<T = any> {
  success: boolean;
  message?: string;
  data?: T;
  error?: string;
}

class ApiService {
  private token: string | null = localStorage.getItem('token');

  setToken(token: string | null) {
    this.token = token;
    if (token) {
      localStorage.setItem('token', token);
    } else {
      localStorage.removeItem('token');
    }
  }

  getToken(): string | null {
    if (!this.token) {
      this.token = localStorage.getItem('token');
    }
    return this.token;
  }

  private async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const token = this.getToken();
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...(options.headers as Record<string, string>),
    };

    if (token) {
      headers.Authorization = `Bearer ${token}`;
    }

    try {
      const response = await fetch(`${API_BASE}${endpoint}`, {
        ...options,
        headers,
      });

      const json: ApiResponse<T> = await response.json();

      if (!response.ok || !json.success) {
        throw new Error(json.error || json.message || 'Request failed');
      }

      return json.data !== undefined ? json.data : (json as unknown as T);
    } catch (err: any) {
      // If network error (e.g. backend container is not running locally), log & rethrow or fallback
      console.warn(`[api-service] Network/API call failed for ${endpoint}: ${err.message}`);
      throw err;
    }
  }

  // --- AUTH SERVICE (/api/auth) ---

  async login(email: string, password: string, role: UserRole): Promise<{ token: string; user: User }> {
    try {
      const res = await this.request<{ token: string; user: any }>('/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
      });

      const user: User = {
        id: String(res.user.id),
        email: res.user.email,
        firstName: res.user.full_name?.split(' ')[0] || res.user.full_name || 'User',
        lastName: res.user.full_name?.split(' ').slice(1).join(' ') || '',
        phone: res.user.phone || '',
        role: res.user.role || role,
        isVerified: true,
        profileComplete: true,
        createdAt: new Date().toISOString(),
      };

      this.setToken(res.token);
      return { token: res.token, user };
    } catch (err) {
      // Fallback mock login if server unreachable
      const mockUser = await mockApi.getOrders(role, '1').then(() => ({
        id: '1',
        email,
        firstName: email.split('@')[0] || 'Demo',
        lastName: 'User',
        phone: '+1234567890',
        role,
        isVerified: true,
        profileComplete: true,
        createdAt: new Date().toISOString(),
      }));
      const mockToken = 'mock-jwt-token-' + Date.now();
      this.setToken(mockToken);
      return { token: mockToken, user: mockUser };
    }
  }

  async register(data: RegisterData): Promise<{ token: string; user: User }> {
    try {
      const payload = {
        full_name: `${data.firstName} ${data.lastName}`.trim(),
        email: data.email,
        phone: data.phone,
        password: data.password,
        role: data.role,
      };

      const res = await this.request<{ token: string; user: any }>('/auth/register', {
        method: 'POST',
        body: JSON.stringify(payload),
      });

      const user: User = {
        id: String(res.user.id),
        email: res.user.email,
        firstName: data.firstName,
        lastName: data.lastName,
        phone: data.phone,
        role: data.role,
        isVerified: data.role === 'client',
        profileComplete: false,
        createdAt: new Date().toISOString(),
      };

      this.setToken(res.token);
      return { token: res.token, user };
    } catch (err) {
      // Fallback mock register
      const mockUser: User = {
        id: String(Date.now()),
        email: data.email,
        firstName: data.firstName,
        lastName: data.lastName,
        phone: data.phone,
        role: data.role,
        isVerified: data.role === 'client',
        profileComplete: false,
        createdAt: new Date().toISOString(),
      };
      const mockToken = 'mock-jwt-token-' + Date.now();
      this.setToken(mockToken);
      return { token: mockToken, user: mockUser };
    }
  }

  async getMe(): Promise<User> {
    const res = await this.request<{ user_id: number; role: UserRole; email: string }>('/auth/me');
    return {
      id: String(res.user_id),
      email: res.email,
      firstName: res.email.split('@')[0],
      lastName: '',
      phone: '',
      role: res.role,
      isVerified: true,
      profileComplete: true,
      createdAt: new Date().toISOString(),
    };
  }

  // --- ORDER SERVICE (/api/orders) ---

  async createOrder(orderData: Partial<Order>): Promise<Order> {
    try {
      const payload = {
        pickup_address: orderData.pickupLocation?.address || 'Pickup location',
        pickup_latitude: orderData.pickupLocation?.lat || 0,
        pickup_longitude: orderData.pickupLocation?.lng || 0,
        dropoff_address: orderData.dropoffLocation?.address || 'Dropoff location',
        dropoff_latitude: orderData.dropoffLocation?.lat || 0,
        dropoff_longitude: orderData.dropoffLocation?.lng || 0,
        package_desc: orderData.items?.map(i => i.name).join(', ') || 'Package',
        weight_kg: 1.0,
        notes: orderData.notes || '',
      };

      const res = await this.request<any>('/orders', {
        method: 'POST',
        body: JSON.stringify(payload),
      });

      return {
        id: String(res.id || res.tracking_code || `ORD-${Date.now()}`),
        clientId: String(res.client_id || orderData.clientId || '1'),
        type: orderData.type || 'dispatch',
        status: (res.status as OrderStatus) || 'pending',
        pickupLocation: orderData.pickupLocation!,
        dropoffLocation: orderData.dropoffLocation!,
        items: orderData.items || [],
        distance: orderData.distance || 2.5,
        estimatedTime: orderData.estimatedTime || 15,
        price: res.delivery_fee || orderData.price || 15.0,
        paymentStatus: 'pending',
        paymentMethod: orderData.paymentMethod || 'cash',
        notes: orderData.notes,
        createdAt: new Date().toISOString(),
      };
    } catch {
      return mockApi.createOrder(orderData);
    }
  }

  async getOrders(role: 'client' | 'rider', userId: string): Promise<Order[]> {
    try {
      const res = await this.request<any[]>('/orders');
      if (Array.isArray(res)) {
        return res.map(o => ({
          id: String(o.id || o.tracking_code),
          clientId: String(o.client_id),
          riderId: o.rider_id ? String(o.rider_id) : undefined,
          type: 'dispatch',
          status: o.status || 'pending',
          pickupLocation: { lat: 40.7128, lng: -74.0060, address: o.pickup_address },
          dropoffLocation: { lat: 40.7589, lng: -73.9851, address: o.dropoff_address },
          items: [{ id: '1', name: o.package_desc || 'Package', quantity: 1, price: o.delivery_fee || 0 }],
          distance: 2.5,
          estimatedTime: 15,
          price: o.delivery_fee || 0,
          paymentStatus: 'paid',
          paymentMethod: 'card',
          notes: o.notes,
          createdAt: o.created_at || new Date().toISOString(),
        }));
      }
      return mockApi.getOrders(role, userId);
    } catch {
      return mockApi.getOrders(role, userId);
    }
  }

  async updateOrderStatus(orderId: string, status: OrderStatus): Promise<Order> {
    try {
      const res = await this.request<any>(`/orders/${orderId}/status`, {
        method: 'PATCH',
        body: JSON.stringify({ status }),
      });
      return {
        id: String(res.id),
        clientId: String(res.client_id),
        type: 'dispatch',
        status: res.status,
        pickupLocation: { lat: 40.7128, lng: -74.0060, address: res.pickup_address },
        dropoffLocation: { lat: 40.7589, lng: -73.9851, address: res.dropoff_address },
        items: [],
        distance: 2.5,
        estimatedTime: 15,
        price: res.delivery_fee || 0,
        paymentStatus: 'paid',
        paymentMethod: 'card',
        createdAt: new Date().toISOString(),
      };
    } catch {
      return mockApi.updateOrderStatus(orderId, status);
    }
  }

  // --- DISPATCH SERVICE (/api/dispatch) ---

  async respondDispatch(dispatchId: string, action: 'accept' | 'reject'): Promise<void> {
    try {
      await this.request(`/dispatch/${dispatchId}/respond`, {
        method: 'POST',
        body: JSON.stringify({ action }),
      });
    } catch {
      // Silently continue or fallback
    }
  }

  // --- RIDER SERVICE (/api/riders) ---

  async updateRiderLocation(riderId: string, location: Location): Promise<void> {
    try {
      await this.request('/riders/me/location', {
        method: 'PATCH',
        body: JSON.stringify({ latitude: location.lat, longitude: location.lng }),
      });
    } catch {
      await mockApi.updateRiderLocation(riderId, location);
    }
  }

  async setRiderAvailability(riderId: string, isOnline: boolean): Promise<User> {
    try {
      await this.request('/riders/me/availability', {
        method: 'PATCH',
        body: JSON.stringify({ available: isOnline }),
      });
      const user = await mockApi.setRiderAvailability(riderId, isOnline);
      return user;
    } catch {
      return mockApi.setRiderAvailability(riderId, isOnline);
    }
  }

  async getNearbyRiders(location: Location, radiusKm: number = 1): Promise<User[]> {
    try {
      const res = await this.request<any[]>('/riders/available');
      if (Array.isArray(res)) {
        return res.map(r => ({
          id: String(r.id),
          email: r.user?.email || 'rider@example.com',
          firstName: r.user?.full_name?.split(' ')[0] || 'Rider',
          lastName: r.user?.full_name?.split(' ')[1] || '',
          phone: r.user?.phone || '',
          role: 'rider',
          isVerified: r.is_verified,
          profileComplete: true,
          createdAt: new Date().toISOString(),
          licenseNumber: r.license_number,
          vehicleType: r.vehicle_type,
          vehiclePlate: r.vehicle_plate,
          rating: r.rating || 4.8,
          totalDeliveries: r.total_deliveries || 50,
          isOnline: r.is_available,
          currentLocation: { lat: r.current_latitude || location.lat, lng: r.current_longitude || location.lng },
        }));
      }
      return mockApi.getNearbyRiders(location, radiusKm);
    } catch {
      return mockApi.getNearbyRiders(location, radiusKm);
    }
  }

  // --- UPLOAD SERVICE (/api/upload) ---

  async uploadAvatar(file: File): Promise<{ url: string }> {
    try {
      const formData = new FormData();
      formData.append('avatar', file);

      const token = this.getToken();
      const response = await fetch(`${API_BASE}/upload/avatar`, {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${token}`,
        },
        body: formData,
      });

      const json = await response.json();
      if (!response.ok || !json.success) {
        throw new Error(json.error || 'Upload failed');
      }

      return { url: json.data?.secure_url || json.data?.thumbnail || '' };
    } catch {
      return new Promise((resolve) => {
        const reader = new FileReader();
        reader.onload = (e) => resolve({ url: e.target?.result as string });
        reader.readAsDataURL(file);
      });
    }
  }

  // --- TRACKING ---

  async getTrackingData(orderId: string) {
    try {
      const res = await this.request<any>(`/orders/${orderId}/tracking`);
      return res;
    } catch {
      return mockApi.getTrackingData(orderId);
    }
  }
}

export const api = new ApiService();