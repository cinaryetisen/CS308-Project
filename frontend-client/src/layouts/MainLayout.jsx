import { Link, Outlet, useNavigate } from 'react-router-dom';
import { useState, useEffect } from 'react';
import Sidebar from '../components/Sidebar';
import ThemeToggle from '../components/ThemeToggle';
import { apiRequest } from '../api/client';

export default function MainLayout() {
    const navigate = useNavigate();

    const [isLoggedIn, setIsLoggedIn]     = useState(false);
    const [showDropdown, setShowDropdown] = useState(false);
    const [userData, setUserData]         = useState(null);
    const [cartCount, setCartCount]       = useState(0);
    const [wishlistCount, setWishlistCount] = useState(0);

    const refreshCartCount = async () => {
        const token = localStorage.getItem("token");
        if (token) {
            try {
                const data = await apiRequest("/api/cart");
                setCartCount(Array.isArray(data) ? data.reduce((sum, item) => sum + item.quantity, 0) : 0);
            } catch {
                // Non-critical: leave count as-is on error
            }
        } else {
            const cart = JSON.parse(localStorage.getItem("cart") || "[]");
            setCartCount(cart.reduce((sum, item) => sum + item.quantity, 0));
        }
    };

    const refreshWishlistCount = async () => {
        const token = localStorage.getItem("token");
        if (!token) return;
        try {
            const data = await apiRequest("/api/wishlist");
            setWishlistCount(Array.isArray(data) ? data.length : 0);
        } catch {
            // Non-critical
        }
    };

    useEffect(() => {
        const token = localStorage.getItem("token");
        if (token) {
            setIsLoggedIn(true);
            apiRequest("/api/users/me")
                .then(user => {
                    if (user) {
                        setUserData(user);
                        localStorage.setItem("user", JSON.stringify(user));
                    }
                })
                .catch(err => console.error("Failed to fetch user profile", err));
        }
        refreshCartCount();
        refreshWishlistCount();
    }, []);

    const handleLogout = () => {
        localStorage.removeItem("token");
        localStorage.removeItem("user");
        localStorage.removeItem("cart");
        setIsLoggedIn(false);
        setShowDropdown(false);
        setUserData(null);
        setCartCount(0);
        navigate('/');
    };

    return (
        <div className="flex flex-col h-screen bg-[var(--bg)] text-[var(--text)] overflow-hidden">

            <header className="shrink-0 w-full z-50 bg-[var(--bg)] border-b border-[var(--border)] px-6 py-4 flex justify-between items-center">

                <Link to="/" className="text-xl font-serif text-[var(--accent)] hover:text-[var(--text)] transition-colors">
                    The Vault of Essence
                </Link>

                <div className="flex items-center gap-4">

                    <ThemeToggle />

                    {/* Wishlist */}
                    <Link to="/wishlist" className="relative hover:scale-110 transition-transform flex items-center justify-center" title="Wishlist">
                        <span className="text-2xl">🤍</span>
                        {wishlistCount > 0 && (
                            <span className="absolute -top-2 -right-2 bg-[var(--accent-dim)] text-[var(--on-accent)] text-[10px] font-bold min-w-[18px] h-[18px] rounded-full flex items-center justify-center shadow-md px-0.5">
                                {wishlistCount > 99 ? "99+" : wishlistCount}
                            </span>
                        )}
                    </Link>

                    {/* Cart */}
                    <Link to="/shoppingcart" className="relative hover:scale-110 transition-transform flex items-center justify-center">
                        <span className="text-2xl">🛒</span>
                        {cartCount > 0 && (
                            <span className="absolute -top-2 -right-2 bg-[var(--accent-dim)] text-[var(--on-accent)] text-[10px] font-bold min-w-[18px] h-[18px] rounded-full flex items-center justify-center shadow-md px-0.5">
                                {cartCount > 99 ? "99+" : cartCount}
                            </span>
                        )}
                    </Link>

                    {isLoggedIn ? (
                        <div className="relative">
                            <button
                                onClick={() => setShowDropdown(!showDropdown)}
                                className="w-10 h-10 rounded-lg bg-[var(--surface-alt)] text-[var(--accent)] font-bold flex items-center justify-center border border-[var(--border)] hover:opacity-80 transition focus:outline-none"
                                title={userData?.name || "Account"}
                            >
                                {userData?.name ? userData.name.charAt(0).toUpperCase() : "U"}
                            </button>

                            {showDropdown && (
                                <div className="absolute right-0 mt-3 w-48 bg-[var(--surface)] border border-[var(--border)] rounded-lg shadow-xl overflow-hidden z-50">
                                    <div className="px-4 py-3 border-b border-[var(--border)]">
                                        <p className="text-xs text-[var(--muted)] mb-0.5">Signed in as</p>
                                        <p className="text-sm font-bold text-[var(--text)] truncate">
                                            {userData?.name || "User"}
                                        </p>
                                    </div>
                                    <Link
                                        to="/profile"
                                        onClick={() => setShowDropdown(false)}
                                        className="block px-4 py-2 text-sm text-[var(--muted)] hover:bg-[var(--surface-alt)] hover:text-[var(--accent)] transition"
                                    >
                                        Profile
                                    </Link>
                                    <Link
                                        to="/orders"
                                        onClick={() => setShowDropdown(false)}
                                        className="block px-4 py-2 text-sm text-[var(--muted)] hover:bg-[var(--surface-alt)] hover:text-[var(--accent)] transition"
                                    >
                                        My Orders
                                    </Link>
                                    <Link
                                        to="/wishlist"
                                        onClick={() => setShowDropdown(false)}
                                        className="block px-4 py-2 text-sm text-[var(--muted)] hover:bg-[var(--surface-alt)] hover:text-[var(--accent)] transition"
                                    >
                                        Wishlist
                                    </Link>
                                    <button
                                        onClick={handleLogout}
                                        className="block w-full text-left px-4 py-2 text-sm text-[#ffb4ab] hover:bg-[var(--surface-alt)] transition"
                                    >
                                        Log Out
                                    </button>
                                </div>
                            )}
                        </div>
                    ) : (
                        <div className="flex items-center gap-3">
                            <Link to="/login" className="text-sm text-[var(--muted)] hover:text-[var(--accent)] transition">
                                Login
                            </Link>
                            <Link to="/signup" className="text-sm font-bold px-4 py-2 bg-[var(--accent)] text-[var(--on-accent)] rounded-lg hover:bg-[var(--accent-dim)] transition">
                                Sign Up
                            </Link>
                        </div>
                    )}
                </div>
            </header>

            <div className="flex flex-1 overflow-hidden">
                <Sidebar />
                <main className="flex-1 overflow-y-auto relative">
                    <Outlet context={{ refreshCartCount, refreshWishlistCount }} />
                </main>
            </div>
        </div>
    );
}
