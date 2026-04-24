import { Link, useNavigate } from 'react-router-dom';
import { useState, useEffect } from 'react';

export default function Payment() {
    const API_URL = import.meta.env.VITE_API_URL;
    const navigate = useNavigate();
    
    const [isProcessing, setIsProcessing] = useState(false);
    const [checkoutError, setCheckoutError] = useState("");
    const [shippingAddress, setShippingAddress] = useState("");

    // Make sure user is logged in
    useEffect(() => {
        const token = localStorage.getItem("token");
        if (!token) {
            navigate("/login");
        }
    }, [navigate]);

    const handlePayment = async (e) => {
        e.preventDefault();
        setIsProcessing(true);
        setCheckoutError("");

        try {
            const token = localStorage.getItem("token");
            if (!token) {
                throw new Error("You must be logged in to complete checkout.");
            }

            // 1. Fetch the REAL cart securely from your Go Backend!
            const cartRes = await fetch(`${API_URL}/api/cart`, {
                method: "GET",
                headers: {
                    "Authorization": `Bearer ${token}`
                }
            });
            
            if (!cartRes.ok) {
                throw new Error("Failed to retrieve cart from server.");
            }

            const cartItems = await cartRes.json();

            // 2. Check if the database cart is actually empty
            if (!cartItems || cartItems.length === 0) {
                setCheckoutError("Your cart is empty.");
                setIsProcessing(false);
                return;
            }

            // 3. Calculate total price from the backend data
            const totalPrice = cartItems.reduce((sum, item) => sum + (item.price * item.quantity), 0);
            
            // 4. Format items for the checkout controller
            // Using the backend 'product_id' and 'quantity' (snake_case) keys
            const formattedCartItems = cartItems.map(item => ({
                product_id: item.product_id, 
                quantity: item.quantity
            }));

            // 5. Build the exact JSON payload your Go controller expects
            const payload = {
                shipping_address: shippingAddress,
                total_price: totalPrice,
                cart_items: formattedCartItems
            };

            // 6. Send request to Go backend
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

            // 7. Success! Clear any local ghost carts and redirect
            localStorage.removeItem("cart");
            alert("Order placed successfully! Your invoice is being dispatched to your email.");
            navigate("/");

        } catch (error) {
            console.error("Checkout Error:", error);
            setCheckoutError(error.message);
        } finally {
            setIsProcessing(false);
        }
    };

    return (
        <div className="min-h-screen flex flex-col items-center py-12 w-full max-w-3xl mx-auto px-6">
            
            {/* Header / Back Link */}
            <div className="w-full mb-8 flex justify-between items-center">
                <Link to="/shoppingcart" className="text-gray-500 dark:text-[#d1c5b0] hover:text-purple-600 dark:hover:text-[#e7b4ff] font-medium transition-colors">
                    &larr; Back to Cart
                </Link>
                <h1 className="text-3xl font-headline italic text-gray-900 dark:text-[#f5ded3]">Secure Checkout</h1>
            </div>

            {/* Payment Form Card */}
            <div className="w-full bg-white dark:bg-[#160c06] p-8 rounded-xl shadow-lg border border-gray-200 dark:border-[#342720]">
                
                {checkoutError && (
                    <div className="mb-6 p-4 bg-red-100 dark:bg-[#93000a] text-red-700 dark:text-[#ffdad6] text-sm rounded-lg border border-red-200 dark:border-transparent text-center font-bold">
                        {checkoutError}
                    </div>
                )}

                <form onSubmit={handlePayment} className="space-y-6">
                    
                    {/* Shipping Address (Required by Go Backend) */}
                    <div>
                        <label className="block text-sm font-label uppercase tracking-widest text-gray-700 dark:text-[#6d6452] mb-2 font-bold">
                            Delivery Address
                        </label>
                        <textarea 
                            required
                            rows="2"
                            value={shippingAddress}
                            onChange={(e) => setShippingAddress(e.target.value)}
                            placeholder="Enter your full shipping address..."
                            className="w-full px-4 py-3 bg-gray-50 dark:bg-[#251912] border border-gray-300 dark:border-[#4e4350] rounded-lg focus:ring-2 focus:ring-purple-500 dark:focus:ring-[#8a47af] focus:outline-none text-gray-900 dark:text-[#f5ded3] placeholder-gray-400 dark:placeholder-[#9a8c9b] transition-colors resize-none"
                        />
                    </div>

                    <div className="h-px w-full bg-gray-200 dark:bg-[#342720] my-6"></div>

                    {/* Cardholder Name */}
                    <div>
                        <label className="block text-sm font-label uppercase tracking-widest text-gray-700 dark:text-[#6d6452] mb-2 font-bold">
                            Cardholder Name
                        </label>
                        <input 
                            type="text" 
                            required
                            placeholder="John Doe"
                            className="w-full px-4 py-3 bg-gray-50 dark:bg-[#251912] border border-gray-300 dark:border-[#4e4350] rounded-lg focus:ring-2 focus:ring-purple-500 dark:focus:ring-[#8a47af] focus:outline-none text-gray-900 dark:text-[#f5ded3] placeholder-gray-400 dark:placeholder-[#9a8c9b] transition-colors"
                        />
                    </div>

                    {/* Card Number */}
                    <div>
                        <label className="block text-sm font-label uppercase tracking-widest text-gray-700 dark:text-[#6d6452] mb-2 font-bold">
                            Card Number
                        </label>
                        <input 
                            type="text" 
                            required
                            maxLength="19"
                            placeholder="0000 0000 0000 0000"
                            className="w-full px-4 py-3 bg-gray-50 dark:bg-[#251912] border border-gray-300 dark:border-[#4e4350] rounded-lg focus:ring-2 focus:ring-purple-500 dark:focus:ring-[#8a47af] focus:outline-none text-gray-900 dark:text-[#f5ded3] placeholder-gray-400 dark:placeholder-[#9a8c9b] transition-colors font-mono tracking-widest"
                        />
                    </div>

                    {/* Expiry and CVC Row */}
                    <div className="flex gap-6">
                        <div className="flex-1">
                            <label className="block text-sm font-label uppercase tracking-widest text-gray-700 dark:text-[#6d6452] mb-2 font-bold">
                                Expiration Date
                            </label>
                            <input 
                                type="text" 
                                required
                                placeholder="MM/YY"
                                maxLength="5"
                                className="w-full px-4 py-3 bg-gray-50 dark:bg-[#251912] border border-gray-300 dark:border-[#4e4350] rounded-lg focus:ring-2 focus:ring-purple-500 dark:focus:ring-[#8a47af] focus:outline-none text-gray-900 dark:text-[#f5ded3] placeholder-gray-400 dark:placeholder-[#9a8c9b] transition-colors text-center"
                            />
                        </div>
                        <div className="flex-1">
                            <label className="block text-sm font-label uppercase tracking-widest text-gray-700 dark:text-[#6d6452] mb-2 font-bold">
                                CVC / CVV
                            </label>
                            <input 
                                type="text" 
                                required
                                placeholder="123"
                                maxLength="4"
                                className="w-full px-4 py-3 bg-gray-50 dark:bg-[#251912] border border-gray-300 dark:border-[#4e4350] rounded-lg focus:ring-2 focus:ring-purple-500 dark:focus:ring-[#8a47af] focus:outline-none text-gray-900 dark:text-[#f5ded3] placeholder-gray-400 dark:placeholder-[#9a8c9b] transition-colors text-center"
                            />
                        </div>
                    </div>

                    {/* Pay Button */}
                    <button 
                        type="submit" 
                        disabled={isProcessing}
                        className={`w-full py-4 mt-6 font-bold text-lg rounded-lg shadow-lg active:scale-95 transition-all duration-150 ${
                            isProcessing 
                            ? "bg-gray-400 dark:bg-[#342720] text-gray-700 dark:text-[#9a8c9b] cursor-not-allowed" 
                            : "bg-gradient-to-r from-[#8a47af] to-[#500075] dark:from-[#e7b4ff] dark:to-[#8a47af] text-white dark:text-[#300049] hover:brightness-110 bubble-pop"
                        }`}
                    >
                        {isProcessing ? "Processing Securely..." : "Confirm Order"}
                    </button>
                    
                </form>
            </div>
        </div>
    );
}