import { createContext, useContext, useState, useEffect, useCallback, ReactNode } from 'react';
import { User, UserRole, AuthState } from '../types';
import { api } from '../services/api';

interface AuthContextType extends AuthState {
  login: (email: string, password: string, role: UserRole) => Promise<void>;
  register: (data: RegisterData) => Promise<void>;
  logout: () => void;
  updateProfile: (data: Partial<User>) => Promise<void>;
  verifyRider: (data: RiderVerificationData) => Promise<void>;
  uploadProfilePicture: (file: File) => Promise<string>;
}

interface RegisterData {
  email: string;
  password: string;
  confirmPassword: string;
  firstName: string;
  lastName: string;
  phone: string;
  role: UserRole;
}

interface RiderVerificationData {
  licenseNumber: string;
  licenseExpiry: string;
  vehicleType: string;
  vehiclePlate: string;
  idDocument: File;
  selfie: File;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const storedUser = localStorage.getItem('user');
    const storedToken = localStorage.getItem('token');
    
    if (storedUser && storedToken) {
      try {
        setUser(JSON.parse(storedUser));
      } catch {
        localStorage.removeItem('user');
        localStorage.removeItem('token');
      }
    }
    setLoading(false);
  }, []);

  const login = useCallback(async (email: string, password: string, role: UserRole) => {
    setLoading(true);
    setError(null);
    
    try {
      const res = await api.login(email, password, role);
      localStorage.setItem('user', JSON.stringify(res.user));
      setUser(res.user);
    } catch (err: any) {
      setError(err.message || 'Invalid credentials');
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  const register = useCallback(async (data: RegisterData) => {
    setLoading(true);
    setError(null);
    
    try {
      const res = await api.register(data);
      localStorage.setItem('user', JSON.stringify(res.user));
      setUser(res.user);
    } catch (err: any) {
      setError(err.message || 'Registration failed');
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  const logout = useCallback(() => {
    localStorage.removeItem('user');
    localStorage.removeItem('token');
    api.setToken(null);
    setUser(null);
  }, []);

  const updateProfile = useCallback(async (data: Partial<User>) => {
    if (!user) return;
    
    setLoading(true);
    try {
      const updatedUser = { ...user, ...data };
      localStorage.setItem('user', JSON.stringify(updatedUser));
      setUser(updatedUser);
    } catch (err: any) {
      setError(err.message || 'Failed to update profile');
      throw err;
    } finally {
      setLoading(false);
    }
  }, [user]);

  const verifyRider = useCallback(async (data: RiderVerificationData) => {
    if (!user || user.role !== 'rider') return;
    
    setLoading(true);
    try {
      const updatedUser = { 
        ...user, 
        isVerified: true,
        licenseNumber: data.licenseNumber,
        vehicleType: data.vehicleType,
        vehiclePlate: data.vehiclePlate,
        profileComplete: true
      };
      
      localStorage.setItem('user', JSON.stringify(updatedUser));
      setUser(updatedUser);
    } catch (err: any) {
      setError(err.message || 'Verification failed');
      throw err;
    } finally {
      setLoading(false);
    }
  }, [user]);

  const uploadProfilePicture = useCallback(async (file: File): Promise<string> => {
    const res = await api.uploadAvatar(file);
    return res.url;
  }, []);

  const value: AuthContextType = {
    user,
    loading,
    error,
    login,
    register,
    logout,
    updateProfile,
    verifyRider,
    uploadProfilePicture
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
}


export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}