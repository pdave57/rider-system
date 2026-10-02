import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Mail, MapPin, Bell, MessageSquare, CheckCircle, AlertCircle, Camera, X, User } from 'lucide-react';
import { useAuth } from '../../contexts/AuthContext';


const profileSteps = [
  { id: 'basic', title: 'Basic Info', icon: User },
  { id: 'address', title: 'Address', icon: MapPin },
  { id: 'preferences', title: 'Preferences', icon: Bell },
];

export function ProfileSetup() {
  const navigate = useNavigate();
  const { user, updateProfile, uploadProfilePicture, loading } = useAuth();
  const [currentStep, setCurrentStep] = useState(0);
  const [avatar, setAvatar] = useState<string | null>(user?.avatar || null);
  const [avatarFile, setAvatarFile] = useState<File | null>(null);
  const [formData, setFormData] = useState({
    firstName: user?.firstName || '',
    lastName: user?.lastName || '',
    email: user?.email || '',
    phone: user?.phone || '',
    address: user?.address || '',
    notifications: user?.preferences?.notifications ?? true,
    smsAlerts: user?.preferences?.smsAlerts ?? false,
    emailAlerts: user?.preferences?.emailAlerts ?? true,
  });
  const [errors, setErrors] = useState<Record<string, string>>({});

  const handleAvatarChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      if (file.size > 5 * 1024 * 1024) {
        setErrors(prev => ({ ...prev, avatar: 'File size must be less than 5MB' }));
        return;
      }
      setAvatarFile(file);
      const reader = new FileReader();
      reader.onload = (event) => setAvatar(event.target?.result as string);
      reader.readAsDataURL(file);
      setErrors(prev => ({ ...prev, avatar: '' }));
    }
  };

  const removeAvatar = () => {
    setAvatar(null);
    setAvatarFile(null);
    setErrors(prev => ({ ...prev, avatar: '' }));
  };

  const validateStep = () => {
    const newErrors: Record<string, string> = {};
    if (currentStep === 0) {
      if (!formData.firstName.trim()) newErrors.firstName = 'First name is required';
      if (!formData.lastName.trim()) newErrors.lastName = 'Last name is required';
      if (!formData.email.trim()) newErrors.email = 'Email is required';
      else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.email)) newErrors.email = 'Invalid email';
      if (!formData.phone.trim()) newErrors.phone = 'Phone is required';
    }
    if (currentStep === 1) {
      if (!formData.address.trim()) newErrors.address = 'Address is required';
    }
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleNext = async () => {
    if (!validateStep()) return;

    if (currentStep === profileSteps.length - 1) {
      try {
        if (avatarFile) {
          const avatarUrl = await uploadProfilePicture(avatarFile);
          formData.avatar = avatarUrl;
        }
        await updateProfile({ 
          ...formData, 
          preferences: {
            notifications: formData.notifications,
            smsAlerts: formData.smsAlerts,
            emailAlerts: formData.emailAlerts,
          },
          profileComplete: true,
        });
        navigate('/client/dashboard');
      } catch {
        setErrors({ submit: 'Failed to save profile' });
      }
    } else {
      setCurrentStep(prev => prev + 1);
    }
  };

  const handleBack = () => {
    if (currentStep > 0) {
      setCurrentStep(prev => prev - 1);
      setErrors({});
    }
  };

  const renderStep = () => {
    switch (profileSteps[currentStep].id) {
      case 'basic':
        return (
          <div className="space-y-4">
            <div className="text-center mb-6">
              <div className="relative inline-block">
                {avatar ? (
                  <img src={avatar} alt="Profile" className="avatar-xl" />
                ) : (
                  <div className="avatar-placeholder-xl text-2xl">
                    {formData.firstName?.[0] || '?'}{formData.lastName?.[0] || ''}
                  </div>
                )}
                <label className="absolute bottom-0 right-0 bg-sky-blue-600 text-white p-2 rounded-full hover:bg-sky-blue-700 transition cursor-pointer">
                  <Camera className="icon" />
                  <input type="file" accept="image/*" onChange={handleAvatarChange} className="sr-only" />
                </label>
                {avatar && (
                  <button type="button" onClick={removeAvatar} className="absolute bottom-0 left-0 bg-red-500 text-white p-2 rounded-full hover:bg-red-600 transition">
                    <X className="icon" />
                  </button>
                )}
              </div>
              {errors.avatar && <p className="text-sm text-red-500 mt-2">{errors.avatar}</p>}
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="form-group">
                <label htmlFor="firstName" className="label">First Name</label>
                <input
                  id="firstName"
                  type="text"
                  className={`input ${errors.firstName ? 'input-error' : ''}`}
                  value={formData.firstName}
                  onChange={(e) => setFormData(prev => ({ ...prev, firstName: e.target.value }))}
                  placeholder="John"
                />
                {errors.firstName && <p className="text-xs text-red-500 mt-1">{errors.firstName}</p>}
              </div>
              <div className="form-group">
                <label htmlFor="lastName" className="label">Last Name</label>
                <input
                  id="lastName"
                  type="text"
                  className={`input ${errors.lastName ? 'input-error' : ''}`}
                  value={formData.lastName}
                  onChange={(e) => setFormData(prev => ({ ...prev, lastName: e.target.value }))}
                  placeholder="Doe"
                />
                {errors.lastName && <p className="text-xs text-red-500 mt-1">{errors.lastName}</p>}
              </div>
            </div>

            <div className="form-group">
              <label htmlFor="email" className="label">Email</label>
              <input
                id="email"
                type="email"
                className={`input ${errors.email ? 'input-error' : ''}`}
                value={formData.email}
                onChange={(e) => setFormData(prev => ({ ...prev, email: e.target.value }))}
                placeholder="john@example.com"
              />
              {errors.email && <p className="text-xs text-red-500 mt-1">{errors.email}</p>}
            </div>

            <div className="form-group">
              <label htmlFor="phone" className="label">Phone Number</label>
              <input
                id="phone"
                type="tel"
                className={`input ${errors.phone ? 'input-error' : ''}`}
                value={formData.phone}
                onChange={(e) => setFormData(prev => ({ ...prev, phone: e.target.value }))}
                placeholder="+1 (555) 000-0000"
              />
              {errors.phone && <p className="text-xs text-red-500 mt-1">{errors.phone}</p>}
            </div>
          </div>
        );

      case 'address':
        return (
          <div className="space-y-4">
            <div className="form-group">
              <label htmlFor="address" className="label">Delivery Address</label>
              <textarea
                id="address"
                rows={3}
                className={`input ${errors.address ? 'input-error' : ''}`}
                value={formData.address}
                onChange={(e) => setFormData(prev => ({ ...prev, address: e.target.value }))}
                placeholder="Enter your full delivery address including apartment/suite number"
              />
              {errors.address && <p className="text-xs text-red-500 mt-1">{errors.address}</p>}
            </div>
            <div className="p-4 bg-gray-50 rounded-lg">
              <p className="text-sm text-gray-600">
                <strong>Tip:</strong> Add landmarks or specific instructions to help riders find you easily.
              </p>
            </div>
          </div>
        );

      case 'preferences':
        return (
          <div className="space-y-4">
            <div className="form-group">
              <div className="flex items-center justify-between">
                <div>
                  <p className="font-medium text-gray-900">Push Notifications</p>
                  <p className="text-sm text-gray-500">Receive order updates and promotions</p>
                </div>
                <button
                  onClick={() => setFormData(prev => ({ ...prev, notifications: !prev.notifications }))}
                  className={`relative w-12 h-7 rounded-full transition ${
                    formData.notifications ? 'bg-sky-blue-600' : 'bg-gray-300'
                  }`}
                  role="switch"
                  aria-checked={formData.notifications}
                >
                  <span className={`absolute top-0.5 transition transform ${
                    formData.notifications ? 'translate-x-6' : 'translate-x-0.5'
                  } w-5 h-5 bg-white rounded-full shadow`} />
                </button>
              </div>
            </div>

            <div className="form-group">
              <div className="flex items-center justify-between">
                <div>
                  <p className="font-medium text-gray-900">SMS Alerts</p>
                  <p className="text-sm text-gray-500">Important updates via text message</p>
                </div>
                <button
                  onClick={() => setFormData(prev => ({ ...prev, smsAlerts: !prev.smsAlerts }))}
                  className={`relative w-12 h-7 rounded-full transition ${
                    formData.smsAlerts ? 'bg-sky-blue-600' : 'bg-gray-300'
                  }`}
                  role="switch"
                  aria-checked={formData.smsAlerts}
                >
                  <span className={`absolute top-0.5 transition transform ${
                    formData.smsAlerts ? 'translate-x-6' : 'translate-x-0.5'
                  } w-5 h-5 bg-white rounded-full shadow`} />
                </button>
              </div>
            </div>

            <div className="form-group">
              <div className="flex items-center justify-between">
                <div>
                  <p className="font-medium text-gray-900">Email Alerts</p>
                  <p className="text-sm text-gray-500">Order confirmations and receipts</p>
                </div>
                <button
                  onClick={() => setFormData(prev => ({ ...prev, emailAlerts: !prev.emailAlerts }))}
                  className={`relative w-12 h-7 rounded-full transition ${
                    formData.emailAlerts ? 'bg-sky-blue-600' : 'bg-gray-300'
                  }`}
                  role="switch"
                  aria-checked={formData.emailAlerts}
                >
                  <span className={`absolute top-0.5 transition transform ${
                    formData.emailAlerts ? 'translate-x-6' : 'translate-x-0.5'
                  } w-5 h-5 bg-white rounded-full shadow`} />
                </button>
              </div>
            </div>
          </div>
        );

      default:
        return null;
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col">
      <div className="bg-white border-b border-gray-200">
        <div className="container py-4">
          <div className="flex items-center justify-between mb-4">
            <h1 className="text-xl font-bold text-navy-blue">Complete Your Profile</h1>
            <span className="text-sm text-gray-500">Step {currentStep + 1} of {profileSteps.length}</span>
          </div>
          <div className="flex items-center gap-2">
            {profileSteps.map((step, index) => (
              <div key={step.id} className="flex items-center flex-1">
                <div
                  className={`w-10 h-10 rounded-full flex items-center justify-center font-medium transition ${
                    index < currentStep
                      ? 'bg-sky-blue-600 text-white'
                      : index === currentStep
                      ? 'bg-sky-blue-100 text-sky-blue-600 border-2 border-sky-blue-200'
                      : 'bg-gray-100 text-gray-400'
                  }`}
                >
                  {index < currentStep ? (
                    <CheckCircle className="icon" />
                  ) : (
                    <span>{index + 1}</span>
                  )}
                </div>
                {index < profileSteps.length - 1 && (
                  <div className={`flex-1 h-1 mx-2 transition ${
                    index < currentStep ? 'bg-sky-blue-600' : 'bg-gray-200'
                  }`} />
                )}
              </div>
            ))}
          </div>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto p-4">
        <div className="container max-w-md">
          <div className="card">
            <div className="card-body">
              {renderStep()}
              {errors.submit && (
                <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg flex items-center gap-2 text-red-700">
                  <AlertCircle className="icon flex-shrink-0" />
                  <span className="text-sm">{errors.submit}</span>
                </div>
              )}
            </div>
            <div className="card-footer flex justify-between">
              <button
                onClick={handleBack}
                disabled={currentStep === 0 || loading}
                className="btn btn-outline"
              >
                Back
              </button>
              <button
                onClick={handleNext}
                disabled={loading}
                className="btn btn-primary"
              >
                {currentStep === profileSteps.length - 1 ? 'Complete Profile' : 'Next'}
                {loading && '...'}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}