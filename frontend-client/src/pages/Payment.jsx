import { Link, useNavigate, useOutletContext } from 'react-router-dom';
import { useState, useEffect } from 'react';

export default function Payment() {
    const API_URL = import.meta.env.VITE_API_URL;
    const navigate = useNavigate();

    // Grab the refresh function passed down from the layout
    const { refreshCartCount } = useOutletContext();

    const [isProcessing, setIsProcessing]   = useState(false);
    const [checkoutError, setCheckoutError] = useState("");
    const [shippingAddress, setShippingAddress] = useState("");

    // Card fields
    const [cardNumber, setCardNumber] = useState("");
    const [expiry, setExpiry]         = useState("");
    const [cvc, setCvc]               = useState("");

    useEffect(() => {
        const token = localStorage.getItem("token");
        if (!token) navigate("/login");
    }, [navigate]);

    // ── Formatters ────────────────────────────────────────────────────────────

    const handleCardNumber = (e) => {
        const digits = e.target.value.replace(/\D/g, "").slice(0, 16);
        const formatted = digits.replace(/(.{4})/g, "$1 ").trim();
        setCardNumber(formatted);
    };

    const handleExpiry = (e) => {
        let digits = e.target.value.replace(/\D/g, "").slice(0, 4);

        // If first digit is 2–9, auto-pad to 0X (e.g. 3 → 03)
        if (digits.length === 1 && parseInt(digits, 10) > 1) {
            digits = "0" + digits;
        }

        // Clamp month to 01–12
        if (digits.length >= 2) {
            const month = parseInt(digits.slice(0, 2), 10);
            if (month < 1 || month > 12) return;
        }

        const formatted = digits.length > 2
            ? `${digits.slice(0, 2)}/${digits.slice(2)}`
            : digits;
        setExpiry(formatted);
    };

    const handleCvc = (e) => {
        setCvc(e.target.value.replace(/\D/g, "").slice(0, 3));
    };

    // ── Submit ────────────────────────────────────────────────────────────────

    const handlePayment = async (e) => {
        e.preventDefault();
        setIsProcessing(true);
        setCheckoutError("");

        try {
            const token = localStorage.getItem("token");
            if (!token) throw new Error("You must be logged in to complete checkout.");

            const cartRes = await fetch(`${API_URL}/api/cart`, {
                method: "GET",
                headers: { "Authorization": `Bearer ${token}` }
            });
            if (!cartRes.ok) throw new Error("Failed to retrieve cart from server.");

            const cartItems = await cartRes.json();
            if (!cartItems || cartItems.length === 0) {
                setCheckoutError("Your cart is empty.");
                setIsProcessing(false);
                return;
            }

            const totalPrice = cartItems.reduce((sum, item) => sum + (item.price * item.quantity), 0);
            const formattedCartItems = cartItems.map(item => ({
                product_id: item.product_id,
                quantity:   item.quantity
            }));

            const payload = {
                shipping_address: shippingAddress,
                total_price:      totalPrice,
                cart_items:       formattedCartItems
            };

            const response = await fetch(`${API_URL}/api/checkout`, {
                method: "POST",
                headers: {
                    "Content-Type":  "application/json",
                    "Authorization": `Bearer ${token}`
                },
                body: JSON.stringify(payload)
            });

            const data = await response.json();
            if (!response.ok) throw new Error(data.error || "Payment failed to process.");

            // Clear local storage and trigger header refresh
            localStorage.removeItem("cart");
            if (refreshCartCount) refreshCartCount();

            // ── F4: Navigate to invoice page instead of alert ──
            navigate(`/invoice/${data.order.delivery_id}`, { state: { order: data.order } });

        } catch (error) {
            console.error("Checkout Error:", error);
            setCheckoutError(error.message);
        } finally {
            setIsProcessing(false);
        }
    };

    const inputClass = "w-full px-4 py-3 bg-gray-50 dark:bg-[#251912] border border-gray-300 dark:border-[#4e4350] rounded-lg focus:ring-2 focus:ring-purple-500 dark:focus:ring-[#8a47af] focus:outline-none text-gray-900 dark:text-[#f5ded3] placeholder-gray-400 dark:placeholder-[#9a8c9b] transition-colors";
    const labelClass = "block text-sm font-label uppercase tracking-widest text-gray-700 dark:text-[#6d6452] mb-2 font-bold";

    return (
        <div className="min-h-screen flex flex-col items-center py-12 w-full max-w-3xl mx-auto px-6">

            {/* Header */}
            <div className="w-full mb-8 flex justify-between items-center">
                <Link to="/shoppingcart" className="text-gray-500 dark:text-[#d1c5b0] hover:text-purple-600 dark:hover:text-[#e7b4ff] font-medium transition-colors">
                    &larr; Back to Cart
                </Link>
                <h1 className="text-3xl font-headline italic text-gray-900 dark:text-[#f5ded3]">Secure Checkout</h1>
            </div>

            {/* Form Card */}
            <div className="w-full bg-white dark:bg-[#160c06] p-8 rounded-xl shadow-lg border border-gray-200 dark:border-[#342720]">

                {checkoutError && (
                    <div className="mb-6 p-4 bg-red-100 dark:bg-[#93000a] text-red-700 dark:text-[#ffdad6] text-sm rounded-lg border border-red-200 dark:border-transparent text-center font-bold">
                        {checkoutError}
                    </div>
                )}

                <form onSubmit={handlePayment} className="space-y-6">

                    {/* Shipping Address */}
                    <div>
                        <label className={labelClass}>Delivery Address</label>
                        <textarea
                            required
                            rows="2"
                            value={shippingAddress}
                            onChange={(e) => setShippingAddress(e.target.value)}
                            placeholder="Enter your full shipping address..."
                            className={`${inputClass} resize-none`}
                        />
                    </div>

                    <div className="h-px w-full bg-gray-200 dark:bg-[#342720] my-6" />

                    {/* Cardholder Name */}
                    <div>
                        <label className={labelClass}>Cardholder Name</label>
                        <input
                            type="text"
                            required
                            placeholder="John Doe"
                            className={inputClass}
                        />
                    </div>

                    {/* Card Number */}
                    <div>
                        <label className={labelClass}>Card Number</label>
                        <input
                            type="text"
                            required
                            inputMode="numeric"
                            value={cardNumber}
                            onChange={handleCardNumber}
                            placeholder="0000 0000 0000 0000"
                            maxLength="19"
                            className={`${inputClass} font-mono tracking-widest`}
                        />
                    </div>

                    {/* Expiry + CVC */}
                    <div className="flex gap-6">
                        <div className="flex-1">
                            <label className={labelClass}>Expiration Date</label>
                            <input
                                type="text"
                                required
                                inputMode="numeric"
                                value={expiry}
                                onChange={handleExpiry}
                                placeholder="MM/YY"
                                maxLength="5"
                                className={`${inputClass} text-center`}
                            />
                        </div>
                        <div className="flex-1">
                            <label className={labelClass}>CVC / CVV</label>
                            <input
                                type="text"
                                required
                                inputMode="numeric"
                                value={cvc}
                                onChange={handleCvc}
                                placeholder="123"
                                maxLength="3"
                                className={`${inputClass} text-center`}
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
