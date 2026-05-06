import { Link, useNavigate } from 'react-router-dom';
import { useState, useEffect } from 'react';

// FIX: Moved API_URL outside the component so it doesn't trigger unnecessary re-renders
const API_URL = import.meta.env.VITE_API_URL;

export default function ShoppingCart() {
    const navigate = useNavigate();

    const [isLoggedIn, setIsLoggedIn]       = useState(false);
    const [cartItems, setCartItems]         = useState([]);
    const [cartLoading, setCartLoading]     = useState(true);
    const [checkoutError, setCheckoutError] = useState("");

    // ── Logic: Fetch Cart ─────────────────────────────────────────────────────
    useEffect(() => {
        const token = localStorage.getItem("token");

        if (token) {
            setIsLoggedIn(true);

            // Fetch cart from backend
            setCartLoading(true);
            fetch(`${API_URL}/api/cart`, {
                headers: { "Authorization": `Bearer ${token}` }
            })
                .then((res) => res.json())
                .then((data) => {
                    setCartItems(
                        (data || []).map((item) => ({
                            id:        item.product_id,
                            name:      item.name,
                            price:     item.price,
                            quantity:  item.quantity,
                            image_url: item.image_url || "",
                            stock:     item.stock || 9999,
                            category:  item.category || "Product"
                        }))
                    );
                })
                .catch(() => setCartItems([]))
                .finally(() => setCartLoading(false));

        } else {
            // Guest — use localStorage
            try {
                setCartItems(JSON.parse(localStorage.getItem("cart") || "[]"));
            } catch {
                setCartItems([]);
            }
            setCartLoading(false);
        }
    }, []); // FIX: Removed API_URL from the dependency array

    // ── Logic: Update Quantity ────────────────────────────────────────────────
    const updateQuantity = async (id, newQty) => {
        if (newQty <= 0) { removeItem(id); return; }

        const currentItem = cartItems.find((item) => item.id === id);
        if (!currentItem) return;
        const delta = newQty - currentItem.quantity;

        // Optimistically update UI first
        const updated = cartItems.map((item) =>
            item.id === id ? { ...item, quantity: newQty } : item
        );
        setCartItems(updated);
        if (!isLoggedIn) { localStorage.setItem("cart", JSON.stringify(updated)); return; }

        // Sync with backend
        try {
            const token = localStorage.getItem("token");
            await fetch(`${API_URL}/api/cart/item`, {
                method: "PATCH",
                headers: {
                    "Content-Type": "application/json",
                    "Authorization": `Bearer ${token}`
                },
                body: JSON.stringify({ product_id: id, quantity: delta })
            });
        } catch (err) {
            console.error("Failed to update quantity:", err);
        }
    };

    // ── Logic: Remove Item ────────────────────────────────────────────────────
    const removeItem = async (id) => {
        // Optimistically update UI first
        const updated = cartItems.filter((item) => item.id !== id);
        setCartItems(updated);
        if (!isLoggedIn) { localStorage.setItem("cart", JSON.stringify(updated)); return; }

        // Sync with backend
        try {
            const token = localStorage.getItem("token");
            await fetch(`${API_URL}/api/cart/${id}`, {
                method: "DELETE",
                headers: { "Authorization": `Bearer ${token}` }
            });
        } catch (err) {
            console.error("Failed to remove item:", err);
        }
    };

    // ── Logic: Calculations & Checkout ────────────────────────────────────────
    // FIX: Added parseInt to guarantee math addition instead of string concatenation
    const totalItems = cartItems.reduce((sum, item) => sum + parseInt(item.quantity || 0, 10), 0);
    const totalCost = cartItems.reduce((sum, item) => sum + item.price * item.quantity, 0);

    const handleCheckout = () => {
        setCheckoutError("");
        if (!isLoggedIn) {
            setCheckoutError("You must be logged in to proceed to payment. Redirecting to login...");
            setTimeout(() => navigate("/login"), 2000);
        } else {
            navigate("/payment");
        }
    };

    // ── Design: UI Render ─────────────────────────────────────────────────────
    return (
        <div className="max-w-screen-xl mx-auto px-6 py-12 grid grid-cols-1 lg:grid-cols-12 gap-12 w-full">
            
            {/* Left: Cart Items */}
            <div className="lg:col-span-8">
                
                {/* Header */}
                <div className="flex items-baseline justify-between mb-8">
                    <h1 className="font-headline text-5xl italic tracking-tight text-gray-900 dark:text-[#f5ded3]">
                        Shopping Cart
                    </h1>
                    <span className="font-label text-xs uppercase tracking-[0.2em] text-gray-500 dark:text-[#6d6452] hidden sm:block">
                        Your Items
                    </span>
                </div>

                {/* Content Rendering */}
                {cartLoading ? (
                    <div className="text-center mt-20">
                        <p className="text-xl text-gray-500 dark:text-[#6d6452] animate-pulse">Loading your cart...</p>
                    </div>
                ) : cartItems.length === 0 ? (
                    <div className="text-center mt-20">
                        <p className="text-xl text-gray-500 dark:text-[#6d6452] mb-6">Your cart is currently empty.</p>
                        <Link to="/">
                            <button className="bg-gradient-to-r from-[#8a47af] to-[#500075] dark:from-[#e7b4ff] dark:to-[#8a47af] text-white dark:text-[#300049] px-8 py-3 rounded-lg font-bold tracking-wider shadow-lg active:scale-95 transition-all duration-150">
                                Browse Products
                            </button>
                        </Link>
                    </div>
                ) : (
                    /* Parent container with divide-y for borders between items */
                    <div className="rounded-lg border border-gray-200 dark:border-[#342720] divide-y divide-gray-200 dark:divide-[#342720] bg-white dark:bg-[#160c06] overflow-hidden shadow-sm">
                        {cartItems.map((item) => (
                            <div 
                                key={item.id} 
                                className="group relative flex flex-col sm:flex-row items-start sm:items-center gap-6 p-6 hover:bg-gray-50 dark:hover:bg-[#251912] transition-colors duration-300"
                            >
                                {/* The transparent light-up hover overlay */}
                                <div className="absolute inset-0 bg-gradient-to-r from-purple-500/10 dark:from-[#e7b4ff]/10 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300 pointer-events-none"></div>
                                
                                <div className="w-24 h-32 flex-shrink-0 bg-gray-100 dark:bg-[#251912] rounded-sm overflow-hidden z-10 mx-auto sm:mx-0 border border-gray-200 dark:border-[#342720]/50">
                                    <img 
                                        className="w-full h-full object-cover" 
                                        src={item.image_url} 
                                        alt={item.name} 
                                    />
                                </div>
                                
                                <div className="flex-grow flex flex-col gap-1 z-10 w-full text-center sm:text-left">
                                    <span className="font-label text-[10px] uppercase tracking-widest text-green-700 dark:text-[#add461]">
                                        {/* FIX: Safely parse category in case backend returns an object */}
                                        {typeof item.category === 'object' && item.category !== null ? item.category.name : (item.category || "Product")}
                                    </span>
                                    <h3 className="font-headline text-2xl text-gray-900 dark:text-[#f5ded3]">{item.name}</h3>
                                    
                                    <button 
                                        onClick={() => removeItem(item.id)}
                                        className="text-red-600 hover:text-red-800 dark:text-[#ffb4ab] dark:hover:text-[#ffdad6] text-xs font-bold uppercase tracking-widest mt-2 sm:mt-1 transition-colors self-center sm:self-start"
                                    >
                                        Remove Item
                                    </button>
                                </div>
                                
                                <div className="flex flex-row sm:flex-col items-center sm:items-end gap-6 sm:gap-4 z-10 w-full sm:w-auto justify-between sm:justify-end mt-4 sm:mt-0">
                                    <div className="flex items-center bg-gray-100 dark:bg-[#342720] rounded px-2 py-1 gap-4 border border-gray-200 dark:border-transparent">
                                        <button 
                                            onClick={() => updateQuantity(item.id, item.quantity - 1)}
                                            className="text-purple-700 dark:text-[#e7b4ff] hover:text-gray-900 dark:hover:text-[#f5ded3] transition-colors flex items-center"
                                        >
                                            <span className="material-symbols-outlined text-lg">-</span>
                                        </button>
                                        <span className="font-bold text-gray-900 dark:text-[#f5ded3] w-4 text-center">
                                            {item.quantity}
                                        </span>
                                        <button 
                                            onClick={() => updateQuantity(item.id, item.quantity + 1)}
                                            className="text-purple-700 dark:text-[#e7b4ff] hover:text-gray-900 dark:hover:text-[#f5ded3] transition-colors flex items-center"
                                        >
                                            <span className="material-symbols-outlined text-lg">+</span>
                                        </button>
                                    </div>
                                    <span className="font-headline text-xl text-gray-900 dark:text-[#f5ded3]">
                                        ${(item.price * item.quantity).toFixed(2)}
                                    </span>
                                </div>
                            </div>
                        ))}
                    </div>
                )}
            </div>

            {/* Right: Order Summary Sticky */}
            {cartItems.length > 0 && !cartLoading && (
                <aside className="lg:col-span-4">
                    <div className="sticky top-28 bg-white dark:bg-[#160c06] p-8 rounded-lg border border-gray-200 dark:border-[#342720] shadow-sm dark:shadow-[0_40px_60px_-15px_rgba(22,12,6,0.6)]">
                        <h2 className="font-headline text-3xl mb-8 border-b border-gray-200 dark:border-[#342720] pb-4 text-gray-900 dark:text-[#f5ded3]">Order Summary</h2>
                        
                        <div className="space-y-4 font-body text-sm text-gray-500 dark:text-[#6d6452] mb-12">
                            <div className="flex justify-between">
                                <span>Subtotal ({totalItems} items)</span>
                                <span className="text-gray-900 dark:text-[#f5ded3]">${totalCost.toFixed(2)}</span>
                            </div>
                        </div>
                        
                        <div className="flex flex-col items-center mb-10 gap-2">
                            <span className="font-label text-xs tracking-widest text-green-700 dark:text-[#add461] uppercase">Total</span>
                            <span className="font-headline text-5xl xl:text-6xl font-bold text-gray-900 dark:text-[#f5ded3] drop-shadow-sm">
                                ${totalCost.toFixed(2)}
                            </span>
                        </div>

                        {checkoutError && (
                            <div className="w-full text-center px-4 py-3 mb-6 text-sm text-white bg-red-600 dark:bg-[#93000a] rounded-md shadow-sm">
                                {checkoutError}
                            </div>
                        )}

                        <div className="flex flex-col gap-4">
                            <button 
                                onClick={handleCheckout}
                                className="bg-gradient-to-r from-[#8a47af] to-[#500075] dark:from-[#e7b4ff] dark:to-[#8a47af] text-white dark:text-[#300049] w-full py-4 font-bold text-lg rounded-lg shadow-lg hover:brightness-110 active:scale-95 transition-all duration-150"
                            >
                                Proceed to Checkout
                            </button>
                            <Link to="/">
                                <button className="active:scale-95 w-full py-4 bg-transparent border border-gray-300 dark:border-[#342720] text-gray-600 dark:text-[#6d6452] font-medium text-sm rounded-lg hover:bg-gray-50 dark:hover:bg-[#251912] hover:text-purple-700 dark:hover:text-[#e7b4ff] transition-all duration-150">
                                    Continue Shopping
                                </button>
                            </Link>
                        </div>

                        <div className="mt-8 flex items-center justify-center gap-2 text-xs text-gray-400 dark:text-[#6d6452]">
                            <span className="material-symbols-outlined text-sm">verified_user</span>
                            <span>Transactions are 100% secure and encrypted</span>
                        </div>
                    </div>
                </aside>
            )}
        </div>
    );
}