import { Link, useNavigate } from 'react-router-dom';
import { useState } from 'react';

export default function Payment() {
    const navigate = useNavigate();
    const [isProcessing, setIsProcessing] = useState(false);

    // Mock processor
    const handlePayment = (e) => {
        e.preventDefault();
        setIsProcessing(true);
        setCheckoutError("");

        try {
            // 1. Retrieve token and cart data
            const token = localStorage.getItem("token");
            const cartItems = JSON.parse(localStorage.getItem("cart") || "[]");

            if (cartItems.length === 0) {
                setCheckoutError("Your cart is empty.");
                setIsProcessing(false);
                return;
            }

            // 2. Calculate total price and format items for Go Backend
            const totalPrice = cartItems.reduce((sum, item) => sum + item.price * item.quantity, 0);
            
            // Map frontend 'id' to backend 'ProductID' expectation
            const formattedCartItems = cartItems.map(item => ({
                ProductID: item.id, 
                Quantity: item.quantity
            }));

            // 3. Build the exact JSON payload your Go controller expects
            const payload = {
                shipping_address: shippingAddress,
                total_price: totalPrice,
                cart_items: formattedCartItems
            };

            // 4. Send request to Go backend
            const response = await fetch(`${API_URL}/api/checkout`, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                    "Authorization": `Bearer ${token}` // Passes JWT to your middleware
                },
                body: JSON.stringify(payload)
            });

            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.error || "Payment failed to process.");
            }

            // 5. Success! Clear cart and redirect
            localStorage.removeItem("cart");
            alert("Order placed successfully! Your invoice is being dispatched to your email.");
            navigate("/");

        } catch (error) {
            console.error("Checkout Error:", error);
            setCheckoutError(error.message);
        } finally {
            setIsProcessing(false);
            alert("Payment Successful! Thank you for your purchase.");
            navigate("/"); // Main store page
        }, 2000);
    };

    return (
        <div className="min-h-screen bg-gray-50 flex flex-col items-center py-10">
            
            {/* Header / Back Link */}
            <div className="w-full max-w-2xl px-6 mb-6 flex justify-between items-center">
                <Link to="/cart" className="text-blue-600 hover:text-blue-800 font-medium transition">
                    &larr; Back to Cart
                </Link>
                <h1 className="text-2xl font-bold text-gray-800">Secure Checkout</h1>
            </div>

            {/* Payment Form Card */}
            <div className="w-full max-w-2xl bg-white p-8 rounded-xl shadow-md border border-gray-200">
                <h2 className="text-xl font-semibold text-gray-800 mb-6">Payment Details</h2>

                <form onSubmit={handlePayment} className="space-y-6">
                    
                    {/* Cardholder Name */}
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Cardholder Name</label>
                        <input 
                            type="text" 
                            required
                            placeholder="John Doe"
                            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:outline-none"
                        />
                    </div>

                    {/* Card Number */}
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Card Number</label>
                        <input 
                            type="text" 
                            required
                            maxLength="19"
                            placeholder="0000 0000 0000 0000"
                            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:outline-none"
                        />
                    </div>

                    {/* Expiry and CVC Row */}
                    <div className="flex gap-6">
                        <div className="flex-1">
                            <label className="block text-sm font-medium text-gray-700 mb-1">Expiration Date</label>
                            <input 
                                type="text" 
                                required
                                placeholder="MM/YY"
                                maxLength="5"
                                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:outline-none"
                            />
                        </div>
                        <div className="flex-1">
                            <label className="block text-sm font-medium text-gray-700 mb-1">CVC / CVV</label>
                            <input 
                                type="text" 
                                required
                                placeholder="123"
                                maxLength="4"
                                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:outline-none"
                            />
                        </div>
                    </div>

                    {/* Pay Button */}
                    <button 
                        type="submit" 
                        disabled={isProcessing}
                        className={`w-full py-3 mt-4 text-white font-bold rounded-lg transition shadow-md ${
                            isProcessing 
                            ? "bg-gray-400 cursor-not-allowed" 
                            : "bg-green-600 hover:bg-green-700"
                        }`}
                    >
                        {isProcessing ? "Processing Securely..." : "Confirm Payment"}
                    </button>
                    
                </form>
            </div>
        </div>
    );
}