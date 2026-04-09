import { Link } from 'react-router-dom';
import { useState } from 'react';

export default function ShoppingCart() {

    // ── Read cart from localStorage ───────────────────────────────────────────
    const [cartItems, setCartItems] = useState(() => {
        try {
            return JSON.parse(localStorage.getItem("cart") || "[]");
        } catch {
            return [];
        }
    });

    // ── Save helper — updates both state and localStorage together ────────────
    const saveCart = (updatedCart) => {
        setCartItems(updatedCart);
        localStorage.setItem("cart", JSON.stringify(updatedCart));
    };

    // ── Update quantity ───────────────────────────────────────────────────────
    const updateQuantity = (id, newQty) => {
        if (newQty <= 0) {
            removeItem(id);
            return;
        }
        const updated = cartItems.map((item) =>
            item.id === id
                ? { ...item, quantity: Math.min(newQty, item.stock) }
                : item
        );
        saveCart(updated);
    };

    // ── Remove item ───────────────────────────────────────────────────────────
    const removeItem = (id) => {
        saveCart(cartItems.filter((item) => item.id !== id));
    };

    // ── Total cost ────────────────────────────────────────────────────────────
    const totalCost = cartItems.reduce(
        (sum, item) => sum + item.price * item.quantity, 0
    );

    return (
        <div className="min-h-screen bg-gray-50 flex flex-col">

            {/* Header */}
            <header className="w-full bg-white shadow-md">
                <div className="max-w-7xl mx-auto px-6 py-4 flex justify-between items-center">
                    <h1 className="text-2xl font-bold text-gray-800">MyStore</h1>
                    <nav className="flex items-center space-x-6">
                        <Link to="/" className="text-gray-600 hover:text-blue-600 transition font-medium">
                            &larr; Back to Shop
                        </Link>
                        <Link to="/login" className="text-gray-700 hover:text-blue-600 transition">
                            Login
                        </Link>
                        <Link to="/signup" className="bg-blue-600 text-white px-5 py-2 rounded-lg hover:bg-blue-700 transition">
                            Sign Up
                        </Link>
                    </nav>
                </div>
            </header>

            {/* Cart Content */}
            <main className="flex-1 max-w-4xl mx-auto w-full px-6 py-10">
                <h2 className="text-3xl font-bold mb-8 text-gray-800 border-b pb-4">
                    Your Shopping Cart
                </h2>

                {cartItems.length === 0 ? (
                    // ── Empty state ───────────────────────────────────────────
                    <div className="text-center mt-20">
                        <p className="text-xl text-gray-500 mb-6">Your cart is currently empty.</p>
                        <Link to="/" className="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 transition">
                            Browse Products
                        </Link>
                    </div>
                ) : (
                    <div className="space-y-6">

                        {/* Item Boxes */}
                        {cartItems.map((item) => (
                            <div
                                key={item.id}
                                className="bg-white border border-gray-200 rounded-xl p-4 flex justify-between items-center shadow-sm gap-4"
                            >
                                {/* Image */}
                                <img
                                    src={item.image_url}
                                    alt={item.name}
                                    className="w-16 h-16 object-contain rounded-lg bg-gray-100"
                                />

                                {/* Name + Quantity controls */}
                                <div className="flex-1">
                                    <h3 className="text-lg font-semibold text-gray-800">{item.name}</h3>
                                    <div className="flex items-center gap-2 mt-2">
                                        <button
                                            onClick={() => updateQuantity(item.id, item.quantity - 1)}
                                            className="w-7 h-7 rounded border border-gray-300 bg-gray-100 hover:bg-gray-200 font-bold text-gray-700"
                                        >
                                            −
                                        </button>
                                        <span className="text-sm font-semibold w-6 text-center">
                                            {item.quantity}
                                        </span>
                                        <button
                                            onClick={() => updateQuantity(item.id, item.quantity + 1)}
                                            className="w-7 h-7 rounded border border-gray-300 bg-gray-100 hover:bg-gray-200 font-bold text-gray-700"
                                        >
                                            +
                                        </button>
                                    </div>
                                </div>

                                {/* Price */}
                                <div className="text-lg font-bold text-gray-800 min-w-[70px] text-right">
                                    ${(item.price * item.quantity).toFixed(2)}
                                </div>

                                {/* Remove */}
                                <button
                                    onClick={() => removeItem(item.id)}
                                    className="text-red-400 hover:text-red-600 text-sm font-medium transition"
                                >
                                    Remove
                                </button>
                            </div>
                        ))}

                        {/* Total + Checkout */}
                        <div className="mt-10 bg-white border border-gray-200 rounded-xl p-6 flex flex-col items-end shadow-sm">
                            <h3 className="text-xl text-gray-600 mb-2">Total</h3>
                            <p className="text-3xl font-bold text-gray-800 mb-6">
                                ${totalCost.toFixed(2)}
                            </p>
                            <Link
                                to="/payment"
                                className="bg-green-600 text-white px-8 py-3 rounded-lg hover:bg-green-700 transition shadow-md text-lg font-bold"
                            >
                                Proceed to Payment
                            </Link>
                        </div>

                    </div>
                )}
            </main>
        </div>
    );
}