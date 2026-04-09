import { Link } from 'react-router-dom';
import { useEffect, useState } from 'react';

export default function Main() {
    const API_URL = import.meta.env.VITE_API_URL;

    const [products, setProducts] = useState([]);
    const [loading, setLoading] = useState(true);
    const [sortOption, setSortOption] = useState("default");

    // Auth state and Dropdown state
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
        // Clear auth data
        localStorage.removeItem("token");
        localStorage.removeItem("user");
        
        // Clear shopping cart
        localStorage.removeItem("cart"); 

        // Reset states
        setIsLoggedIn(false);
        setShowDropdown(false);
        setUserData(null);
    };

    // Fetch products
    useEffect(() => {
        const fetchProducts = async () => {
            try {
                const response = await fetch(`${API_URL}/api/products`);
                const data = await response.json();
                setProducts(data);
            } catch (error) {
                console.error("Error fetching products:", error);
            } finally {
                setLoading(false);
            }
        };

        fetchProducts();
    }, [API_URL]);

    const addToCart = (product) => {
        const cart = JSON.parse(localStorage.getItem("cart") || "[]");
        const existingItemIndex = cart.findIndex((item) => item.id === product.id);

        if (existingItemIndex >= 0) {
            const currentQty = cart[existingItemIndex].quantity;
            cart[existingItemIndex].quantity = Math.min(currentQty + 1, product.quantity);
        } else {
            cart.push({
                id: product.id,
                name: product.name,
                price: product.price,
                image_url: product.image_url,
                stock: product.quantity, 
                quantity: 1 
            });
        }

        localStorage.setItem("cart", JSON.stringify(cart));
        alert(`${product.name} added to cart!`); 
    };

    const sortedProducts = [...products].sort((a, b) => {
        if (sortOption === "price-asc") return a.price - b.price;
        if (sortOption === "price-desc") return b.price - a.price;
        if (sortOption === "rating-desc") return b.rating - a.rating;
        if (sortOption === "name-asc") return a.name.localeCompare(b.name);
        return 0; 
    });

    return (
        <div className="min-h-screen bg-gray-50 flex flex-col">

            {/* Header */}
            <header className="w-full bg-white shadow-md relative z-50">
                <div className="max-w-7xl mx-auto px-6 py-4 flex justify-between items-center">
                    <h1 className="text-2xl font-bold text-gray-800">
                        MyStore
                    </h1>
                    <nav className="flex items-center space-x-6">
                        
                        <Link to="/shoppingcart" className="text-gray-700 hover:text-blue-600 transition">
                            Shopping Cart
                        </Link>

                        {/* Conditional Rendering based on Auth State */}
                        {isLoggedIn ? (
                            <div className="relative">
                                {/* Dropdown Toggle Button */}
                                <button 
                                    onClick={() => setShowDropdown(!showDropdown)}
                                    className="flex items-center space-x-2 text-gray-700 hover:text-blue-600 transition focus:outline-none"
                                >
                                    <span className="font-medium">
                                        {userData?.name || "My Account"}
                                    </span>
                                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M19 9l-7 7-7-7"></path>
                                    </svg>
                                </button>

                                {/* Dropdown Menu */}
                                {showDropdown && (
                                    <div className="absolute right-0 mt-3 w-48 bg-white border border-gray-200 rounded-lg shadow-lg overflow-hidden z-50">
                                        <div className="px-4 py-3 border-b border-gray-100">
                                            <p className="text-sm text-gray-500">Signed in as</p>
                                            <p className="text-sm font-bold text-gray-900 truncate">
                                                {userData?.name || "User"}
                                            </p>
                                            <p className="text-xs text-gray-400 truncate mt-0.5">
                                                {userData?.email}
                                            </p>
                                        </div>
                                        <Link 
                                            to="/profile" 
                                            className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 transition"
                                        >
                                            Profile
                                        </Link>
                                        <Link 
                                            to="/orders" 
                                            className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 transition"
                                        >
                                            My Orders
                                        </Link>
                                        <button 
                                            onClick={handleLogout}
                                            className="block w-full text-left px-4 py-2 text-sm text-red-600 hover:bg-gray-100 transition"
                                        >
                                            Log Out
                                        </button>
                                    </div>
                                )}
                            </div>
                        ) : (
                            <>
                                <Link to="/login" className="text-gray-700 hover:text-blue-600 transition">
                                    Login
                                </Link>
                                <Link to="/signup" className="bg-blue-600 text-white px-5 py-2 rounded-lg hover:bg-blue-700 transition">
                                    Sign Up
                                </Link>
                            </>
                        )}
                    </nav>
                </div>
            </header>

            {/* Main Content */}
            <main className="flex-1 max-w-7xl mx-auto w-full px-6 py-10">

                <div className="flex flex-col md:flex-row justify-between items-center mb-10">
                    <h2 className="text-4xl font-bold text-center md:text-left mb-4 md:mb-0">
                        Our Products
                    </h2>

                    <div className="flex items-center space-x-2">
                        <label htmlFor="sort" className="text-gray-700 font-medium">Sort by:</label>
                        <select 
                            id="sort" 
                            value={sortOption} 
                            onChange={(e) => setSortOption(e.target.value)}
                            className="p-2 border border-gray-300 rounded-lg bg-white shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                        >
                            <option value="default">Featured</option>
                            <option value="price-asc">Price: Low to High</option>
                            <option value="price-desc">Price: High to Low</option>
                            <option value="rating-desc">Highest Rated</option>
                            <option value="name-asc">Name: A to Z</option>
                        </select>
                    </div>
                </div>

                {loading ? (
                    <p className="text-center text-gray-600">Loading products...</p>
                ) : (
                    <div className="grid gap-8 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
                        {sortedProducts.map((product) => (
                            <div
                                key={product.id}
                                className="bg-white rounded-2xl shadow-md p-6 flex flex-col hover:shadow-xl transition duration-300"
                            >
                                <div className="w-full h-56 bg-gray-200 rounded-xl mb-5 flex items-center justify-center">
                                    <img
                                        src={product.image_url}
                                        alt={product.name}
                                        className="h-full object-contain"
                                    />
                                </div>

                                <h3 className="text-xl font-semibold mb-2">
                                    {product.name}
                                </h3>

                                <div className="flex items-center mb-2">
                                    <span className="text-yellow-500 mr-2">
                                        {"★".repeat(Math.floor(product.rating || 0))}
                                    </span>
                                    <span className="text-gray-600 text-sm">
                                        {product.rating} ({product.review_count || 0})
                                    </span>
                                </div>

                                <p className="text-lg font-medium text-gray-800 mb-2">
                                    ${product.price}
                                </p>

                                <p className={`text-sm mb-4 ${
                                    product.quantity === 0 ? "text-red-500" : "text-green-600"
                                }`}>
                                    {product.quantity === 0
                                        ? "Out of Stock"
                                        : `${product.quantity} left in stock`}
                                </p>

                                <button
                                    onClick={() => addToCart(product)} 
                                    disabled={product.quantity === 0}
                                    className={`mt-auto px-4 py-2 rounded-lg transition ${
                                        product.quantity === 0
                                            ? "bg-gray-300 cursor-not-allowed"
                                            : "bg-blue-600 text-white hover:bg-blue-700"
                                    }`}
                                >
                                    {product.quantity === 0 ? "Unavailable" : "Buy Now"}
                                </button>
                            </div>
                        ))}
                    </div>
                )}
            </main>
        </div>
    );
}