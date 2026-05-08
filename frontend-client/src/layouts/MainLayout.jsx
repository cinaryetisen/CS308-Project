import { Link, Outlet, useNavigate } from 'react-router-dom';
import { useState, useEffect } from 'react';
import Sidebar from '../components/Sidebar';

export default function MainLayout() {
    const navigate = useNavigate();
    const API_URL = import.meta.env.VITE_API_URL;

    const [isLoggedIn, setIsLoggedIn] = useState(false);
    const [showDropdown, setShowDropdown] = useState(false);
    const [userData, setUserData] = useState(null);
    const [cartCount, setCartCount] = useState(0);

    const refreshCartCount = async () => {
        const token = localStorage.getItem("token");
        if (token) {
            try {
                const res = await fetch(`${API_URL}/api/cart`, {
                    headers: { Authorization: `Bearer ${token}` },
                });
                if (res.ok) {
                    const data = await res.json();
                    setCartCount(Array.isArray(data) ? data.reduce((sum, item) => sum + item.quantity, 0) : 0);
                }
            } catch (err) {
                console.error("Failed to fetch cart count", err);
            }
        } else {
            const cart = JSON.parse(localStorage.getItem("cart") || "[]");
            setCartCount(cart.reduce((sum, item) => sum + item.quantity, 0));
        }
    };

    useEffect(() => {
        const token = localStorage.getItem("token");

        if (token) {
            setIsLoggedIn(true);

            // Always fetch fresh user data from the API so the name is always up to date
            fetch(`${API_URL}/api/users/me`, {
                headers: { Authorization: `Bearer ${token}` }
            })
                .then(res => res.ok ? res.json() : null)
                .then(user => {
                    if (user) {
                        setUserData(user);
                        localStorage.setItem("user", JSON.stringify(user));
                    }
                })
                .catch(err => console.error("Failed to fetch user profile", err));
        }

        refreshCartCount();
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
        <div className="flex flex-col h-screen bg-[#1c110b] text-[#f5ded3] overflow-hidden">

            <header className="shrink-0 w-full z-50 bg-[#1c110b] border-b border-[#342720] px-6 py-4 flex justify-between items-center">

                <h2 className="text-xl font-serif text-[#e7b4ff]">
                    The Vault
                </h2>

                <div className="flex gap-4">
                    <Link to="/" className="px-4 py-2 bg-[#342720] text-[#e7b4ff] rounded-lg hover:bg-[#40322a] transition">
                        Shop
                    </Link>
                </div>

                <div className="flex items-center gap-6">
                    <Link to="/shoppingcart" className="text-2xl hover:scale-110 transition-transform">
                        🛒 {cartCount > 0 && (
                        <span className="relative -top-8 -right-4 bg-purple-600 text-white text-[10px] font-bold w-5 h-5 rounded-full flex items-center justify-center shadow-md">
                                {cartCount > 99 ? "99+" : cartCount}
                            </span>
                    )}
                    </Link>

                    {isLoggedIn ? (
                        <div className="relative">
                            <button
                                onClick={() => setShowDropdown(!showDropdown)}
                                className="w-10 h-10 rounded-lg bg-[#342720] text-[#e7b4ff] font-bold flex items-center justify-center border border-[#4e4350] hover:bg-[#40322a] transition focus:outline-none"
                                title={userData?.name || "Account"}
                            >
                                {userData?.name ? userData.name.charAt(0).toUpperCase() : "U"}
                            </button>

                            {showDropdown && (
                                <div className="absolute right-0 mt-3 w-48 bg-[#251912] border border-[#342720] rounded-lg shadow-xl overflow-hidden z-50">
                                    <div className="px-4 py-3 border-b border-[#342720]">
                                        <p className="text-xs text-[#d1c5b0] mb-0.5">Signed in as</p>
                                        <p className="text-sm font-bold text-[#f5ded3] truncate">
                                            {userData?.name || "User"}
                                        </p>
                                    </div>
                                    <Link
                                        to="/profile"
                                        onClick={() => setShowDropdown(false)}
                                        className="block px-4 py-2 text-sm text-[#d1c5b0] hover:bg-[#342720] hover:text-[#e7b4ff] transition"
                                    >
                                        Profile
                                    </Link>
                                    <Link
                                        to="/orders"
                                        onClick={() => setShowDropdown(false)}
                                        className="block px-4 py-2 text-sm text-[#d1c5b0] hover:bg-[#342720] hover:text-[#e7b4ff] transition"
                                    >
                                        My Orders
                                    </Link>
                                    <button
                                        onClick={handleLogout}
                                        className="block w-full text-left px-4 py-2 text-sm text-[#ffb4ab] hover:bg-[#342720] transition"
                                    >
                                        Log Out
                                    </button>
                                </div>
                            )}
                        </div>
                    ) : (
                        <div className="flex items-center gap-3">
                            <Link to="/login" className="text-sm text-[#d1c5b0] hover:text-[#e7b4ff] transition">
                                Login
                            </Link>
                            <Link to="/signup" className="text-sm font-bold px-4 py-2 bg-[#69258e] text-[#f6d9ff] rounded-lg hover:bg-[#8a47af] transition">
                                Sign Up
                            </Link>
                        </div>
                    )}
                </div>
            </header>

            <div className="flex flex-1 overflow-hidden">
                <Sidebar />
                <main className="flex-1 overflow-y-auto relative">
                    <Outlet context={{ refreshCartCount }} />
                </main>
            </div>
        </div>
    );
}