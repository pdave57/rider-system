import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Mail, Lock, User, Phone, Eye, EyeOff, AlertCircle, CheckCircle } from 'lucide-react';
import { useAuth } from '../../contexts/AuthContext';

export function Register() {
  const navigate = useNavigate();
  const { register, loading, error } = useAuth();
  const [formData, setFormData] = useState({
    firstName: '',
    lastName: '',
    email: '',
    phone: '',
    password: '',
    confirmPassword: '',
  });
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [formError, setFormError] = useState('');
  const [validations, setValidations] = useState({
    firstName: true,
    lastName: true,
    email: true,
    phone: true,
    password: true,
    confirmPassword: true,
  });

  const validateField = (name: string, value: string) => {
    const newValidations = { ...validations };
    switch (name) {
      case 'firstName':
      case 'lastName':
        newValidations[name] = value.length >= 2;
        break;
      case 'email':
        newValidations[name] = /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value);
        break;
      case 'phone':
        newValidations[name] = /^\+?[\d\s-]{10,}$/.test(value);
        break;
      case 'password':
        newValidations[name] = value.length >= 8;
        break;
      case 'confirmPassword':
        newValidations[name] = value === formData.password && value.length > 0;
        break;
    }
    setValidations(newValidations);
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
    validateField(name, value);
    if (name === 'password') {
      validateField('confirmPassword', formData.confirmPassword);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormError('');

    const isValid = Object.values(validations).every(v => v) && 
      Object.values(formData).every(v => v.trim());

    if (!isValid) {
      setFormError('Please fill in all fields correctly');
      Object.keys(formData).forEach(key => validateField(key, formData[key]));
      return;
    }

    try {
      await register({
        ...formData,
        role: 'client',
      });
      navigate('/client/profile-setup');
    } catch {
      setFormError(error || 'Registration failed');
    }
  };

  const passwordRequirements = [
    { label: 'At least 8 characters', met: formData.password.length >= 8 },
    { label: 'One uppercase letter', met: /[A-Z]/.test(formData.password) },
    { label: 'One lowercase letter', met: /[a-z]/.test(formData.password) },
    { label: 'One number', met: /\d/.test(formData.password) },
  ];

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-navy-blue mb-4">
            <svg className="icon-xl text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728A9 9 0 015.636 5.636m12.728 12.728L5.636 5.636" />
            </svg>
          </div>
          <h1 className="text-3xl font-bold text-navy-blue">Create Account</h1>
          <p className="text-gray-500 mt-2">Join as a Client</p>
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
              <div className="grid grid-cols-2 gap-4 mb-4">
                <div className="form-group">
                  <label htmlFor="firstName" className="label">First Name</label>
                  <div className="relative">
                    <User className="icon absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
                    <input
                      id="firstName"
                      name="firstName"
                      type="text"
                      className={`input pl-10 ${!validations.firstName && formData.firstName ? 'input-error' : ''}`}
                      value={formData.firstName}
                      onChange={handleChange}
                      placeholder="John"
                      autoComplete="given-name"
                      disabled={loading}
                    />
                  </div>
                  {!validations.firstName && formData.firstName && (
                    <p className="text-xs text-red-500 mt-1">At least 2 characters</p>
                  )}
                </div>
                <div className="form-group">
                  <label htmlFor="lastName" className="label">Last Name</label>
                  <div className="relative">
                    <User className="icon absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
                    <input
                      id="lastName"
                      name="lastName"
                      type="text"
                      className={`input pl-10 ${!validations.lastName && formData.lastName ? 'input-error' : ''}`}
                      value={formData.lastName}
                      onChange={handleChange}
                      placeholder="Doe"
                      autoComplete="family-name"
                      disabled={loading}
                    />
                  </div>
                  {!validations.lastName && formData.lastName && (
                    <p className="text-xs text-red-500 mt-1">At least 2 characters</p>
                  )}
                </div>
              </div>

              <div className="form-group">
                <label htmlFor="email" className="label">Email Address</label>
                <div className="relative">
                  <Mail className="icon absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
                  <input
                    id="email"
                    name="email"
                    type="email"
                    className={`input pl-10 ${!validations.email && formData.email ? 'input-error' : ''}`}
                    value={formData.email}
                    onChange={handleChange}
                    placeholder="you@example.com"
                    autoComplete="email"
                    disabled={loading}
                  />
                </div>
                {!validations.email && formData.email && (
                  <p className="text-xs text-red-500 mt-1">Enter a valid email</p>
                )}
              </div>

              <div className="form-group">
                <label htmlFor="phone" className="label">Phone Number</label>
                <div className="relative">
                  <Phone className="icon absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
                  <input
                    id="phone"
                    name="phone"
                    type="tel"
                    className={`input pl-10 ${!validations.phone && formData.phone ? 'input-error' : ''}`}
                    value={formData.phone}
                    onChange={handleChange}
                    placeholder="+1 (555) 000-0000"
                    autoComplete="tel"
                    disabled={loading}
                  />
                </div>
                {!validations.phone && formData.phone && (
                  <p className="text-xs text-red-500 mt-1">Enter a valid phone number</p>
                )}
              </div>

              <div className="form-group">
                <label htmlFor="password" className="label">Password</label>
                <div className="relative">
                  <Lock className="icon absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
                  <input
                    id="password"
                    name="password"
                    type={showPassword ? 'text' : 'password'}
                    className={`input pl-10 pr-10 ${!validations.password && formData.password ? 'input-error' : ''}`}
                    value={formData.password}
                    onChange={handleChange}
                    placeholder="••••••••"
                    autoComplete="new-password"
                    disabled={loading}
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600"
                  >
                    {showPassword ? <EyeOff className="icon" /> : <Eye className="icon" />}
                  </button>
                </div>
                {!validations.password && formData.password && (
                  <p className="text-xs text-red-500 mt-1">At least 8 characters</p>
                )}
                <div className="mt-2 space-y-1">
                  {passwordRequirements.map((req, i) => (
                    <div key={i} className="flex items-center gap-2 text-xs">
                      {req.met ? (
                        <CheckCircle className="icon text-green-500 flex-shrink-0" />
                      ) : (
                        <svg className="icon text-gray-300 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                      )}
                      <span className={req.met ? 'text-green-600' : 'text-gray-500'}>{req.label}</span>
                    </div>
                  ))}
                </div>
              </div>

              <div className="form-group">
                <label htmlFor="confirmPassword" className="label">Confirm Password</label>
                <div className="relative">
                  <Lock className="icon absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
                  <input
                    id="confirmPassword"
                    name="confirmPassword"
                    type={showConfirmPassword ? 'text' : 'password'}
                    className={`input pl-10 pr-10 ${!validations.confirmPassword && formData.confirmPassword ? 'input-error' : ''}`}
                    value={formData.confirmPassword}
                    onChange={handleChange}
                    placeholder="••••••••"
                    autoComplete="new-password"
                    disabled={loading}
                  />
                  <button
                    type="button"
                    onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600"
                  >
                    {showConfirmPassword ? <EyeOff className="icon" /> : <Eye className="icon" />}
                  </button>
                </div>
                {!validations.confirmPassword && formData.confirmPassword && (
                  <p className="text-xs text-red-500 mt-1">Passwords do not match</p>
                )}
              </div>

              <div className="flex items-start gap-2 mb-6">
                <input
                  type="checkbox"
                  id="terms"
                  required
                  className="mt-1 w-4 h-4 rounded border-gray-300 text-sky-blue-600 focus:ring-sky-blue-500"
                />
                <label htmlFor="terms" className="text-sm text-gray-600">
                  I agree to the{' '}
                  <a href="#" className="text-sky-blue-600 hover:underline">Terms of Service</a>{' '}
                  and{' '}
                  <a href="#" className="text-sky-blue-600 hover:underline">Privacy Policy</a>
                </label>
              </div>

              <button
                type="submit"
                className="btn btn-primary btn-block btn-lg"
                disabled={loading}
              >
                {loading ? 'Creating account...' : 'Create Account'}
              </button>
            </form>
          </div>
          <div className="card-footer flex justify-center">
            <p className="text-gray-600">
              Already have an account?{' '}
              <button onClick={() => navigate('/login')} className="text-sky-blue-600 font-medium hover:underline">
                Sign in
              </button>
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}