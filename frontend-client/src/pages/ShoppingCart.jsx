import { Link, useNavigate, useOutletContext } from 'react-router-dom';
import { useState, useEffect } from 'react';
import { apiRequest } from '../api/client';

export default function ShoppingCart() {
    const navigate = useNavigate();

    // Grab the refresh function passed down from MainLayout
    const { refreshCartCount } = useOutletContext();
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
            apiRequest("/api/cart")
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
    }, []);

    // ── Logic: Update Quantity ────────────────────────────────────────────────
    const updateQuantity = async (id, newQty) => {
        if (newQty <= 0) { removeItem(id); return; }

        const currentItem = cartItems.find((item) => item.id === id);
        if (!currentItem) return;

        if (newQty > currentItem.stock) {
            return;
        }

        const delta = newQty - currentItem.quantity;

        // Optimistically update UI first
        const updated = cartItems.map((item) =>
            item.id === id ? { ...item, quantity: newQty } : item
        );
        setCartItems(updated);

        if (!isLoggedIn) {
            localStorage.setItem("cart", JSON.stringify(updated));
            refreshCartCount();
            return;
        }

        // Sync with backend
        try {
            await apiRequest("/api/cart/item", {
                method: "PATCH",
                body: JSON.stringify({ product_id: id, quantity: delta })
            });
            refreshCartCount();
        } catch (err) {
            console.error("Failed to update quantity:", err);
            setCheckoutError(err.message);
        }
    };

    // ── Logic: Remove Item ────────────────────────────────────────────────────
    const removeItem = async (id) => {
        // Optimistically update UI first
        const updated = cartItems.filter((item) => item.id !== id);
        setCartItems(updated);

        if (!isLoggedIn) {
            localStorage.setItem("cart", JSON.stringify(updated));
            refreshCartCount();
            return;
        }

        // Sync with backend
        try {
            await apiRequest(`/api/cart/${id}`, { method: "DELETE" });
            refreshCartCount();
        } catch (err) {
            console.error("Failed to remove item:", err);
            setCheckoutError(err.message);
        }
    };

    // ── Logic: Calculations & Checkout ────────────────────────────────────────
    const totalItems = cartItems.reduce((sum, item) => sum + parseInt(item.quantity || 0, 10), 0);
    const totalCost = cartItems.reduce((sum, item) => sum + item.price * item.quantity, 0);

    const handleCheckout = () => {
        setCheckoutError("");
        if (!isLoggedIn) {
            setCheckoutError("You must be logged in to proceed to payment.");
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
                    <h1 className="font-headline text-5xl italic tracking-tight text-[var(--text)]">
                        Shopping Cart
                    </h1>
                    <span className="font-label text-xs uppercase tracking-[0.2em] text-[var(--muted)] hidden sm:block">
                        Your Items
                    </span>
                </div>

                {/* Content Rendering */}
                {cartLoading ? (
                    <div className="text-center mt-20">
                        <p className="text-xl text-[var(--muted)] animate-pulse">Loading your cart...</p>
                    </div>
                ) : cartItems.length === 0 ? (
                    <div className="text-center mt-20">
                        <p className="text-xl text-[var(--muted)] mb-6">Your cart is currently empty.</p>
                        <Link to="/">
                            <button className="bg-gradient-to-r from-[var(--btn-from)] to-[var(--btn-to)] text-[var(--on-accent)] px-8 py-3 rounded-lg font-bold tracking-wider shadow-lg active:scale-95 transition-all duration-150">
                                Browse Products
                            </button>
                        </Link>
                    </div>
                ) : (
                    /* Parent container with divide-y for borders between items */
                    <div className="rounded-lg border border-[var(--border)] divide-y divide-[var(--border)] bg-[var(--surface)] overflow-hidden shadow-sm">
                        {cartItems.map((item) => (
                            <div
                                key={item.id}
                                className="group relative flex flex-col sm:flex-row items-start sm:items-center gap-6 p-6 hover:bg-[var(--surface-alt)] transition-colors duration-300"
                            >
                                {/* The transparent light-up hover overlay */}
                                <div className="absolute inset-0 bg-gradient-to-r from-[var(--accent)]/10 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300 pointer-events-none"></div>

                                <div className="w-24 h-32 flex-shrink-0 bg-[var(--surface-alt)] rounded-sm overflow-hidden z-10 mx-auto sm:mx-0 border border-[var(--border)]/50">
                                    <img
                                        className="w-full h-full object-cover"
                                        src={item.image_url}
                                        alt={item.name}
                                    />
                                </div>

                                <div className="flex-grow flex flex-col gap-1 z-10 w-full text-center sm:text-left">
                                    <span className="font-label text-[10px] uppercase tracking-widest text-[#add461]">
                                        {typeof item.category === 'object' && item.category !== null ? item.category.name : (item.category || "Product")}
                                    </span>
                                    <h3 className="font-headline text-2xl text-[var(--text)]">{item.name}</h3>

                                    <button
                                        onClick={() => removeItem(item.id)}
                                        className="text-[#ffb4ab] hover:text-[#ffdad6] text-xs font-bold uppercase tracking-widest mt-2 sm:mt-1 transition-colors self-center sm:self-start"
                                    >
                                        Remove Item
                                    </button>
                                </div>

                                <div className="flex flex-row sm:flex-col items-center sm:items-end gap-6 sm:gap-4 z-10 w-full sm:w-auto justify-between sm:justify-end mt-4 sm:mt-0">
                                    <div className="flex items-center bg-[var(--surface-alt)] rounded px-2 py-1 gap-4 border border-[var(--border)]/50">
                                        <button
                                            onClick={() => updateQuantity(item.id, item.quantity - 1)}
                                            className="text-[var(--accent)] hover:text-[var(--text)] transition-colors flex items-center"
                                        >
                                            <span className="material-symbols-outlined text-lg">-</span>
                                        </button>
                                        <span className="font-bold text-[var(--text)] w-4 text-center">
                                            {item.quantity}
                                        </span>
                                        <button
                                            onClick={() => updateQuantity(item.id, item.quantity + 1)}
                                            disabled={item.quantity >= item.stock}
                                            title={item.quantity >= item.stock ? "Maximum stock reached" : "Increase quantity"}
                                            className={`transition-colors flex items-center ${
                                                item.quantity >= item.stock
                                                    ? "text-[var(--border)] cursor-not-allowed"
                                                    : "text-[var(--accent)] hover:text-[var(--text)]"
                                            }`}
                                        >
                                            <span className="material-symbols-outlined text-lg">+</span>
                                        </button>
                                    </div>
                                    <span className="font-headline text-xl text-[var(--text)]">
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
                    <div className="sticky top-28 bg-[var(--surface)] p-8 rounded-lg border border-[var(--border)] shadow-sm">
                        <h2 className="font-headline text-3xl mb-8 border-b border-[var(--border)] pb-4 text-[var(--text)]">Order Summary</h2>

                        <div className="space-y-4 font-body text-sm text-[var(--muted)] mb-12">
                            <div className="flex justify-between">
                                <span>Subtotal ({totalItems} items)</span>
                                <span className="text-[var(--text)]">${totalCost.toFixed(2)}</span>
                            </div>
                        </div>

                        <div className="flex flex-col items-center mb-10 gap-2">
                            <span className="font-label text-xs tracking-widest text-[#add461] uppercase">Total</span>
                            <span className="font-headline text-5xl xl:text-6xl font-bold text-[var(--text)] drop-shadow-sm">
                                ${totalCost.toFixed(2)}
                            </span>
                        </div>

                        {checkoutError && (
                            <div className="w-full text-center px-4 py-3 mb-6 text-sm text-[#ffdad6] bg-[#93000a] rounded-md shadow-sm">
                                {checkoutError}
                            </div>
                        )}

                        <div className="flex flex-col gap-4">
                            <button
                                onClick={handleCheckout}
                                className="bg-gradient-to-r from-[var(--btn-from)] to-[var(--btn-to)] text-[var(--on-accent)] w-full py-4 font-bold text-lg rounded-lg shadow-lg hover:brightness-110 active:scale-95 transition-all duration-150"
                            >
                                Proceed to Checkout
                            </button>
                            <Link to="/">
                                <button className="active:scale-95 w-full py-4 bg-transparent border border-[var(--border)] text-[var(--muted)] font-medium text-sm rounded-lg hover:bg-[var(--surface-alt)] hover:text-[var(--accent)] transition-all duration-150">
                                    Continue Shopping
                                </button>
                            </Link>
                        </div>

                        <div className="mt-8 flex items-center justify-center gap-2 text-xs text-[var(--muted)]">
                            <span className="material-symbols-outlined text-sm">verified_user</span>
                            <span>Transactions are 100% secure and encrypted</span>
                        </div>
                    </div>
                </aside>
            )}
        </div>
    );
}
