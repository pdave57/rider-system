import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Mail, Lock, Eye, EyeOff, AlertCircle } from 'lucide-react';
import { useAuth } from '../../contexts/AuthContext';


export function Login() {
  const navigate = useNavigate();
  const { login, loading, error } = useAuth();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [formError, setFormError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormError('');
    
    if (!email || !password) {
      setFormError('Please fill in all fields');
      return;
    }

    try {
      await login(email, password, 'client');
      navigate('/client/dashboard');
    } catch {
      setFormError(error || 'Login failed');
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-navy-blue mb-4">
            <svg className="icon-xl text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728A9 9 0 015.636 5.636m12.728 12.728L5.636 5.636" />
            </svg>
          </div>
          <h1 className="text-3xl font-bold text-navy-blue">Welcome Back</h1>
          <p className="text-gray-500 mt-2">Sign in to your account</p>
        </div>

        <div className="card">
          <div className="card-body">
            {formError && (
              <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg flex items-center gap-2 text-red-700">
                <AlertCircle className="icon flex-shrink-0" />
                <span className="text-sm">{formError}</span>
              </div>
            )}

            <form onSubmit={handleSubmit}>
              <div className="form-group">
                <label htmlFor="email" className="label">Email Address</label>
                <div className="relative">
                  <Mail className="icon absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
                  <input
                    id="email"
                    type="email"
                    className="input pl-10"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    placeholder="you@example.com"
                    autoComplete="email"
                    disabled={loading}
                  />
                </div>
              </div>

              <div className="form-group">
                <label htmlFor="password" className="label">Password</label>
                <div className="relative">
                  <Lock className="icon absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
                  <input
                    id="password"
                    type={showPassword ? 'text' : 'password'}
                    className="input pl-10 pr-10"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    placeholder="••••••••"
                    autoComplete="current-password"
                    disabled={loading}
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600"
                    aria-label={showPassword ? 'Hide password' : 'Show password'}
                  >
                    {showPassword ? <EyeOff className="icon" /> : <Eye className="icon" />}
                  </button>
                </div>
              </div>

              <div className="flex items-center justify-between mb-6">
                <label className="flex items-center gap-2 cursor-pointer">
                  <input type="checkbox" className="w-4 h-4 rounded border-gray-300 text-sky-blue-600 focus:ring-sky-blue-500" />
                  <span className="text-sm text-gray-600">Remember me</span>
                </label>
                <a href="#" className="text-sm text-sky-blue-600 hover:underline">Forgot password?</a>
              </div>

              <button
                type="submit"
                className="btn btn-primary btn-block btn-lg"
                disabled={loading}
              >
                {loading ? 'Signing in...' : 'Sign In'}
              </button>
            </form>
          </div>
          <div className="card-footer flex justify-center">
            <p className="text-gray-600">
              Don't have an account?{' '}
              <button onClick={() => navigate('/role-select')} className="text-sky-blue-600 font-medium hover:underline">
                Sign up
              </button>
            </p>
          </div>
        </div>

        <p className="text-center text-sm text-gray-500 mt-6">
          By continuing, you agree to our{' '}
          <a href="#" className="text-sky-blue-600 hover:underline">Terms of Service</a>{' '}
          and{' '}
          <a href="#" className="text-sky-blue-600 hover:underline">Privacy Policy</a>
        </p>
      </div>
    </div>
  );
}