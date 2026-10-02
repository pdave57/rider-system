import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { User, Bike, Shield, Star, MapPin, Clock, DollarSign, Package, ShoppingBag } from 'lucide-react';


interface RoleOption {
  value: 'client' | 'rider';
  title: string;
  subtitle: string;
  icon: React.ReactNode;
  features: { icon: React.ReactNode; text: string }[];
  color: string;
  bgColor: string;
}

const roles: RoleOption[] = [
  {
    value: 'client',
    title: 'Client',
    subtitle: 'Order deliveries & shop with ease',
    icon: (
      <div className="w-16 h-16 rounded-2xl bg-sky-blue-100 flex items-center justify-center">
        <User className="icon-xl text-sky-blue-600" />
      </div>
    ),
    features: [
      { icon: <Package className="icon" />, text: 'Order for dispatch - Send packages anywhere' },
      { icon: <ShoppingBag className="icon" />, text: 'Shop for me - Groceries & essentials delivered' },
      { icon: <MapPin className="icon" />, text: 'Real-time tracking - Watch your order live' },
      { icon: <Clock className="icon" />, text: 'Fast delivery - Riders matched within 1km' },
    ],
    color: 'text-sky-blue-600',
    bgColor: 'bg-sky-blue-50',
  },
  {
    value: 'rider',
    title: 'Rider',
    subtitle: 'Earn money delivering orders',
    icon: (
      <div className="w-16 h-16 rounded-2xl bg-green-100 flex items-center justify-center">
        <Bike className="icon-xl text-green-600" />
      </div>
    ),
    features: [
      { icon: <DollarSign className="icon" />, text: 'Flexible earnings - Work on your schedule' },
      { icon: <MapPin className="icon" />, text: 'Nearby orders - Matched within 1km radius' },
      { icon: <Shield className="icon" />, text: 'Identity verification - Secure platform' },
      { icon: <Star className="icon" />, text: 'Build rating - More orders with high ratings' },
    ],
    color: 'text-green-600',
    bgColor: 'bg-green-50',
  },
];

export function RoleSelect() {
  const navigate = useNavigate();
  const [selectedRole, setSelectedRole] = useState<'client' | 'rider' | null>(null);

  const handleContinue = () => {
    if (selectedRole) {
      navigate(`/register?role=${selectedRole}`);
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col">
      <div className="flex-1 flex items-center justify-center p-4">
        <div className="w-full max-w-2xl">
          <div className="text-center mb-10">
            <div className="inline-flex items-center justify-center w-20 h-20 rounded-2xl bg-navy-blue mb-6">
              <svg className="icon-2xl text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728A9 9 0 015.636 5.636m12.728 12.728L5.636 5.636" />
              </svg>
            </div>
            <h1 className="text-3xl font-bold text-navy-blue mb-2">Who are you?</h1>
            <p className="text-gray-500 text-lg">Choose your role to get started</p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {roles.map((role) => (
              <button
                key={role.value}
                onClick={() => setSelectedRole(role.value)}
                className={`relative p-6 rounded-2xl border-2 transition-all duration-200 ${
                  selectedRole === role.value
                    ? `border-2 ${role.color} ${role.bgColor} shadow-lg shadow-${role.color.replace('text-', '')}/20`
                    : 'border-gray-200 bg-white hover:border-gray-300 hover:shadow-md'
                }`}
              >
                <div className="absolute top-4 right-4">
                  {selectedRole === role.value ? (
                    <div className="w-6 h-6 rounded-full bg-sky-blue-600 border-2 border-white flex items-center justify-center">
                      <svg className="icon text-white" fill="currentColor" viewBox="0 0 20 20"><path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" /></svg>
                    </div>
                  ) : (
                    <div className="w-6 h-6 rounded-full border-2 border-gray-300" />
                  )}
                </div>

                <div className="mb-4">{role.icon}</div>
                
                <h3 className="text-xl font-bold text-navy-blue mb-1">{role.title}</h3>
                <p className="text-gray-500 mb-6">{role.subtitle}</p>

                <ul className="space-y-3">
                  {role.features.map((feature, i) => (
                    <li key={i} className="flex items-start gap-3">
                      <div className={`flex-shrink-0 w-8 h-8 rounded-lg ${role.bgColor} flex items-center justify-center`}>
                        {feature.icon}
                      </div>
                      <span className="text-sm text-gray-600 mt-0.5">{feature.text}</span>
                    </li>
                  ))}
                </ul>
              </button>
            ))}
          </div>
        </div>
      </div>

      <div className="p-4 border-t border-gray-200 bg-white">
        <button
          onClick={handleContinue}
          disabled={!selectedRole}
          className="btn btn-primary btn-block btn-lg"
        >
          Continue as {selectedRole ? selectedRole.charAt(0).toUpperCase() + selectedRole.slice(1) : 'Select Role'}
        </button>
      </div>
    </div>
  );
}