import { useState, useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { MapPin, Package, ShoppingBag, X, ArrowLeft, CreditCard, DollarSign, Plus, Minus, Trash2 } from 'lucide-react';
import { useOrders } from '../../hooks/useOrders';
import { useAuth } from '../../contexts/AuthContext';

interface OrderItem {
  id: string;
  name: string;
  quantity: number;
  price: number;
  notes?: string;
}

export function NewOrder() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { user } = useAuth();
  const { createOrder, loading: creating } = useOrders(user?.id || null, 'client');
  
  const orderType = (searchParams.get('type') as 'dispatch' | 'shop-for-me') || 'dispatch';
  
  const [step, setStep] = useState(1);
  const [pickupAddress, setPickupAddress] = useState('');
  const [pickupDetails, setPickupDetails] = useState('');
  const [dropoffAddress, setDropoffAddress] = useState('');
  const [dropoffDetails, setDropoffDetails] = useState('');
  const [items, setItems] = useState<OrderItem[]>([{ id: '1', name: '', quantity: 1, price: 0 }]);
  const [notes, setNotes] = useState('');
  const [paymentMethod, setPaymentMethod] = useState<'cash' | 'card' | 'wallet'>('card');
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [calculatedPrice, setCalculatedPrice] = useState(0);
  const [distance, setDistance] = useState(0);

  useEffect(() => {
    const basePrice = orderType === 'dispatch' ? 5 : 3;
    const itemTotal = items.reduce((sum, item) => sum + item.price * item.quantity, 0);
    const distanceFee = distance * 1.5;
    const total = basePrice + itemTotal + distanceFee;
    setCalculatedPrice(Math.round(total * 100) / 100);
  }, [items, distance, orderType]);

  const validateStep1 = () => {
    const newErrors: Record<string, string> = {};
    if (!pickupAddress.trim()) newErrors.pickupAddress = 'Pickup address is required';
    if (!dropoffAddress.trim()) newErrors.dropoffAddress = 'Dropoff address is required';
    if (pickupAddress === dropoffAddress && pickupAddress) newErrors.dropoffAddress = 'Addresses must be different';
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const validateStep2 = () => {
    if (orderType === 'shop-for-me') {
      const newErrors: Record<string, string> = {};
      const hasValidItems = items.some(item => item.name.trim() && item.quantity > 0);
      if (!hasValidItems) newErrors.items = 'At least one item is required';
      setErrors(newErrors);
      return Object.keys(newErrors).length === 0;
    }
    return true;
  };

  const handleNext = () => {
    if (step === 1 && validateStep1()) {
      setStep(2);
    } else if (step === 2 && validateStep2()) {
      setStep(3);
    }
  };

  const handleBack = () => {
    setStep(prev => Math.max(1, prev - 1));
    setErrors({});
  };

  const handleSubmit = async () => {
    if (!validateStep2()) return;
    
    setErrors({});
    try {
      await createOrder({
        clientId: user?.id,
        type: orderType,
        pickupLocation: { 
          lat: 40.7128 + (Math.random() - 0.5) * 0.01, 
          lng: -74.0060 + (Math.random() - 0.5) * 0.01,
          address: pickupAddress,
        },
        dropoffLocation: { 
          lat: 40.7128 + (Math.random() - 0.5) * 0.01, 
          lng: -74.0060 + (Math.random() - 0.5) * 0.01,
          address: dropoffAddress,
        },
        items: items.filter(i => i.name.trim()),
        distance: distance || 2.5,
        estimatedTime: Math.round((distance || 2.5) * 5 + items.length * 3),
        price: calculatedPrice,
        paymentMethod,
        notes,
      });
      navigate('/client/orders');
    } catch (err) {
      setErrors({ submit: 'Failed to create order. Please try again.' });
    }
  };

  const addItem = () => {
    setItems(prev => [...prev, { id: Date.now().toString(), name: '', quantity: 1, price: 0 }]);
  };

  const removeItem = (id: string) => {
    if (items.length > 1) {
      setItems(prev => prev.filter(item => item.id !== id));
    }
  };

  const updateItem = (id: string, field: keyof OrderItem, value: any) => {
    setItems(prev => prev.map(item => item.id === id ? { ...item, [field]: value } : item));
  };

  const progressSteps = [
    { label: 'Locations', number: 1 },
    { label: orderType === 'dispatch' ? 'Details' : 'Items', number: 2 },
    { label: 'Payment', number: 3 },
  ];

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col">
      <div className="bg-white border-b border-gray-200">
        <div className="container py-4">
          <div className="flex items-center justify-between mb-6">
            <button onClick={() => navigate(-1)} className="p-2 rounded-lg hover:bg-gray-100 transition">
              <ArrowLeft className="icon text-gray-600" />
            </button>
            <h1 className="text-xl font-bold text-navy-blue flex-1 text-center">
              {orderType === 'dispatch' ? 'New Dispatch Order' : 'Shop for Me'}
            </h1>
            <div className="w-10" />
          </div>

          <div className="flex items-center gap-2 mb-4">
            {progressSteps.map((s, i) => (
              <div key={s.label} className="flex items-center flex-1">
                <div
                  className={`w-8 h-8 rounded-full flex items-center justify-center font-medium transition ${
                    i < step - 1
                      ? 'bg-sky-blue-600 text-white'
                      : i === step - 1
                      ? 'bg-sky-blue-100 text-sky-blue-600 border-2 border-sky-blue-200'
                      : 'bg-gray-100 text-gray-400'
                  }`}
                >
                  {i < step - 1 ? <X className="icon-sm" /> : s.number}
                </div>
                {i < progressSteps.length - 1 && (
                  <div className={`flex-1 h-1 mx-2 transition ${
                    i < step - 1 ? 'bg-sky-blue-600' : 'bg-gray-200'
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
              {errors.submit && (
                <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">
                  {errors.submit}
                </div>
              )}

              {step === 1 && (
                <div className="space-y-4">
                  <h3 className="text-lg font-semibold text-navy-blue">Pickup & Dropoff Locations</h3>
                  
                  <div className="form-group">
                    <label className="label flex items-center gap-2">
                      <MapPin className="icon text-sky-blue-600" />
                      Pickup Address
                    </label>
                    <textarea
                      rows={2}
                      className={`input ${errors.pickupAddress ? 'input-error' : ''}`}
                      value={pickupAddress}
                      onChange={(e) => setPickupAddress(e.target.value)}
                      placeholder="Enter pickup address (restaurant, store, sender location)"
                    />
                    {errors.pickupAddress && <p className="text-xs text-red-500 mt-1">{errors.pickupAddress}</p>}
                  </div>

                  <div className="form-group">
                    <label className="label">Pickup Details (Optional)</label>
                    <input
                      type="text"
                      className="input"
                      value={pickupDetails}
                      onChange={(e) => setPickupDetails(e.target.value)}
                      placeholder="Contact person, phone, specific instructions"
                    />
                  </div>

                  <div className="form-group">
                    <label className="label flex items-center gap-2">
                      <MapPin className="icon text-red-500" />
                      Dropoff Address
                    </label>
                    <textarea
                      rows={2}
                      className={`input ${errors.dropoffAddress ? 'input-error' : ''}`}
                      value={dropoffAddress}
                      onChange={(e) => setDropoffAddress(e.target.value)}
                      placeholder="Enter delivery address"
                    />
                    {errors.dropoffAddress && <p className="text-xs text-red-500 mt-1">{errors.dropoffAddress}</p>}
                  </div>

                  <div className="form-group">
                    <label className="label">Dropoff Details (Optional)</label>
                    <input
                      type="text"
                      className="input"
                      value={dropoffDetails}
                      onChange={(e) => setDropoffDetails(e.target.value)}
                      placeholder="Contact person, phone, leave at door, etc."
                    />
                  </div>

                  <div className="p-3 bg-gray-50 rounded-lg">
                    <p className="text-sm text-gray-600">
                      <strong>Distance:</strong> ~{distance || 2.5} km • 
                      <strong>Est. Time:</strong> {Math.round((distance || 2.5) * 5)} min
                    </p>
                  </div>
                </div>
              )}

              {step === 2 && orderType === 'dispatch' && (
                <div className="space-y-4">
                  <h3 className="text-lg font-semibold text-navy-blue">Package Details</h3>
                  
                  <div className="form-group">
                    <label className="label">What are you sending?</label>
                    <input
                      type="text"
                      className="input"
                      placeholder="Documents, clothes, electronics, etc."
                      value={items[0]?.name}
                      onChange={(e) => updateItem(items[0].id, 'name', e.target.value)}
                    />
                  </div>

                  <div className="form-group">
                    <label className="label">Special Instructions</label>
                    <textarea
                      rows={3}
                      className="input"
                      placeholder="Fragile, handle with care, requires signature, etc."
                      value={notes}
                      onChange={(e) => setNotes(e.target.value)}
                    />
                  </div>
                </div>
              )}

              {step === 2 && orderType === 'shop-for-me' && (
                <div className="space-y-4">
                  <div className="flex items-center justify-between">
                    <h3 className="text-lg font-semibold text-navy-blue">Shopping List</h3>
                    <button onClick={addItem} className="btn btn-outline btn-sm">
                      <Plus className="icon-sm" />
                      Add Item
                    </button>
                  </div>

                  <div className="space-y-3">
                    {items.map((item, index) => (
                      <div key={item.id} className="flex items-center gap-2 p-3 bg-gray-50 rounded-lg">
                        <input
                          type="number"
                          min="1"
                          max="99"
                          className="w-16 input text-center"
                          value={item.quantity}
                          onChange={(e) => updateItem(item.id, 'quantity', parseInt(e.target.value) || 1)}
                        />
                        <input
                          type="text"
                          className="flex-1 input"
                          placeholder="Item name"
                          value={item.name}
                          onChange={(e) => updateItem(item.id, 'name', e.target.value)}
                        />
                        <input
                          type="number"
                          min="0"
                          step="0.01"
                          className="w-24 input text-right"
                          placeholder="$0.00"
                          value={item.price}
                          onChange={(e) => updateItem(item.id, 'price', parseFloat(e.target.value) || 0)}
                        />
                        {items.length > 1 && (
                          <button
                            onClick={() => removeItem(item.id)}
                            className="p-2 text-red-500 hover:bg-red-50 rounded"
                          >
                            <Trash2 className="icon" />
                          </button>
                        )}
                      </div>
                    ))}
                  </div>

                  <div className="form-group">
                    <label className="label">Special Instructions for Shopper</label>
                    <textarea
                      rows={3}
                      className="input"
                      placeholder="Preferred brands, substitutes if out of stock, etc."
                      value={notes}
                      onChange={(e) => setNotes(e.target.value)}
                    />
                  </div>
                </div>
              )}

              {step === 3 && (
                <div className="space-y-4">
                  <h3 className="text-lg font-semibold text-navy-blue">Payment</h3>
                  
                  <div className="space-y-2">
                    {['card', 'wallet', 'cash'].map((method) => (
                      <label key={method} className={`flex items-center gap-3 p-3 rounded-lg border-2 transition cursor-pointer ${
                        paymentMethod === method
                          ? 'border-sky-blue-500 bg-sky-blue-50'
                          : 'border-gray-200 hover:border-gray-300'
                      }`}>
                        <input
                          type="radio"
                          name="payment"
                          value={method}
                          checked={paymentMethod === method}
                          onChange={() => setPaymentMethod(method as any)}
                          className="w-4 h-4 text-sky-blue-600 border-gray-300 focus:ring-sky-blue-500"
                        />
                        <div className="flex items-center gap-3 flex-1">
                          <div className="w-10 h-10 rounded-lg bg-gray-100 flex items-center justify-center">
                            {method === 'card' && <CreditCard className="icon text-gray-600" />}
                            {method === 'wallet' && <DollarSign className="icon text-green-600" />}
                            {method === 'cash' && <Package className="icon text-gray-600" />}
                          </div>
                          <span className="font-medium text-gray-900 capitalize">{method === 'shop-for-me' ? 'Wallet' : method}</span>
                        </div>
                      </label>
                    ))}
                  </div>

                  <div className="p-4 bg-gray-50 rounded-lg space-y-2">
                    <div className="flex justify-between text-sm">
                      <span className="text-gray-600">Base fare ({orderType === 'dispatch' ? 'Dispatch' : 'Shopping'})</span>
                      <span className="font-medium">${orderType === 'dispatch' ? 5 : 3}.00</span>
                    </div>
                    <div className="flex justify-between text-sm">
                      <span className="text-gray-600">Distance fee (~{distance || 2.5} km)</span>
                      <span className="font-medium">${(distance * 1.5).toFixed(2)}</span>
                    </div>
                    {orderType === 'shop-for-me' && items.some(i => i.price > 0) && (
                      <div className="flex justify-between text-sm">
                        <span className="text-gray-600">Estimated items cost</span>
                        <span className="font-medium">${items.reduce((sum, i) => sum + i.price * i.quantity, 0).toFixed(2)}</span>
                      </div>
                    )}
                    <div className="border-t border-gray-200 pt-2 flex justify-between">
                      <span className="font-semibold text-gray-900">Total</span>
                      <span className="font-bold text-lg text-navy-blue">${calculatedPrice.toFixed(2)}</span>
                    </div>
                  </div>
                </div>
              )}
            </div>
            <div className="card-footer flex justify-between">
              <button
                onClick={handleBack}
                disabled={step === 1 || creating}
                className="btn btn-outline"
              >
                Back
              </button>
              <button
                onClick={step === 3 ? handleSubmit : handleNext}
                disabled={creating}
                className="btn btn-primary"
              >
                {step === 3 ? 'Place Order' : 'Next'}
                {creating && '...'}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}