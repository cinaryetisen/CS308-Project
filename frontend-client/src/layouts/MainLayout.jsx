import { Link, Outlet, useNavigate } from 'react-router-dom';
import { useState, useEffect } from 'react';

export default function MainLayout() {
    const navigate = useNavigate();

    // ── Auth & Dropdown State ───────────────────────────────────────────────
    const [isLoggedIn, setIsLoggedIn] = useState(false);
    const [showDropdown, setShowDropdown] = useState(false);
    const [userData, setUserData] = useState(null);

    // Check for token on component mount
    useEffect(() => {
        const token = localStorage.getItem("token");
        const user = localStorage.getItem("user");
        
        if (token) {
            setIsLoggedIn(true);
            if (user) {
                setUserData(JSON.parse(user));
            }
        }
    }, []);

    // Handle Logout
    const handleLogout = () => {
        localStorage.removeItem("token");
        localStorage.removeItem("user");
        localStorage.removeItem("cart"); 

        setIsLoggedIn(false);
        setShowDropdown(false);
        setUserData(null);
        
        navigate('/login');
    };

    return (
        <div className="min-h-screen bg-[#1c110b] text-[#f5ded3] flex flex-col">

            {/* TOP HEADER (full width) */}
            <header className="w-full sticky top-0 z-50 bg-[#1c110b] border-b border-[#342720] px-6 py-4 flex justify-between items-center">
                
                {/* Logo / Brand */}
                <h2 className="text-xl font-serif text-[#e7b4ff]">
                    The Vault
                </h2>

                {/* Center Links */}
                <div className="flex gap-4">
                    <Link to="/" className="px-4 py-2 bg-[#342720] text-[#e7b4ff] rounded-lg hover:bg-[#40322a] transition">
                        Shop
                    </Link>
                    <Link to="/shoppingcart" className="px-4 py-2 bg-[#342720] text-[#e7b4ff] rounded-lg hover:bg-[#40322a] transition">
                        Cart
                    </Link>
                </div>

                {/* Right Side Icons & Menu */}
                <div className="flex items-center gap-6">
                    <Link to="/shoppingcart" className="text-2xl hover:scale-110 transition-transform">
                        🛒
                    </Link>
                    
                    {/* User Menu / Auth Logic */}
                    {isLoggedIn ? (
                        <div className="relative">
                            {/* The Menu Button (Replaces your empty w-8 h-8 div) */}
                            <button 
                                onClick={() => setShowDropdown(!showDropdown)}
                                className="w-10 h-10 rounded-lg bg-[#342720] text-[#e7b4ff] font-bold flex items-center justify-center border border-[#4e4350] hover:bg-[#40322a] transition focus:outline-none"
                                title={userData?.name || "Account"}
                            >
                                {/* Show the first letter of the user's name, or a default 'U' */}
                                {userData?.name ? userData.name.charAt(0).toUpperCase() : "U"}
                            </button>

                            {/* The Dropdown Panel */}
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
                        /* Logged Out State: Show Login/Signup */
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

            {/* Page Content Injected Here */}
            <main className="flex-1 flex flex-col">
                <Outlet />
            </main>


            

        </div>
    );
}