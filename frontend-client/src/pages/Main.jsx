import { Link } from 'react-router-dom';
import { useEffect, useState } from 'react';

export default function Main() {
    const API_URL = import.meta.env.VITE_API_URL;

    const [products, setProducts] = useState([]);
    const [loading, setLoading] = useState(true);
    const [sortOption, setSortOption] = useState("default");

    // 🔍 NEW: Search state
    const [searchQuery, setSearchQuery] = useState("");

    // Auth state and Dropdown state
    const [isLoggedIn, setIsLoggedIn] = useState(false);
    const [showDropdown, setShowDropdown] = useState(false);
    const [userData, setUserData] = useState(null);

    // Check auth
    useEffect(() => {
        const token = localStorage.getItem("token");
        const user = localStorage.getItem("user");

        if (token) {
            setIsLoggedIn(true);
            if (user) setUserData(JSON.parse(user));
        }
    }, []);

    // Logout
    const handleLogout = () => {
        localStorage.removeItem("token");
        localStorage.removeItem("user");
        localStorage.removeItem("cart");
        setIsLoggedIn(false);
        setShowDropdown(false);
        setUserData(null);
    };

    // 🔥 FETCH with search + sort
    useEffect(() => {
        const fetchProducts = async () => {
            setLoading(true);
            try {
                let url = `${API_URL}/api/products?`;

                if (searchQuery) {
                    url += `search=${encodeURIComponent(searchQuery)}&`;
                }

                if (sortOption !== "default") {
                    const sortMap = {
                        "price-asc": "price_asc",
                        "price-desc": "price_desc",
                        "rating-desc": "popular",
                        "name-asc": "name_asc"
                    };

                    url += `sort=${sortMap[sortOption]}`;
                }

                const response = await fetch(url);
                const data = await response.json();
                setProducts(data);

            } catch (error) {
                console.error("Error fetching products:", error);
            } finally {
                setLoading(false);
            }
        };

        fetchProducts();
    }, [API_URL, searchQuery, sortOption]);

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

    return (
        <div className="min-h-screen bg-gray-50 flex flex-col">

            {/* Header */}
            <header className="w-full bg-white shadow-md relative z-50">
                <div className="max-w-7xl mx-auto px-6 py-4 flex items-center">

                    {/* Logo */}
                    <h1 className="text-2xl font-bold text-gray-800">
                        MyStore
                    </h1>

                    {/* 🔍 SEARCH BAR */}
                    <div className="flex-1 mx-6">
                        <input
                            type="text"
                            placeholder="Search products..."
                            value={searchQuery}
                            onChange={(e) => setSearchQuery(e.target.value)}
                            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                        />
                    </div>

                    {/* Navigation */}
                    <nav className="flex items-center space-x-6">

                        <Link to="/shoppingcart" className="text-gray-700 hover:text-blue-600 transition">
                            Shopping Cart
                        </Link>

                        {isLoggedIn ? (
                            <div className="relative">
                                <button
                                    onClick={() => setShowDropdown(!showDropdown)}
                                    className="flex items-center space-x-2 text-gray-700 hover:text-blue-600 transition"
                                >
                                    <span className="font-medium">
                                        {userData?.name || "My Account"}
                                    </span>
                                </button>

                                {showDropdown && (
                                    <div className="absolute right-0 mt-3 w-48 bg-white border rounded-lg shadow-lg">
                                        <Link to="/profile" className="block px-4 py-2 hover:bg-gray-100">
                                            Profile
                                        </Link>
                                        <Link to="/orders" className="block px-4 py-2 hover:bg-gray-100">
                                            My Orders
                                        </Link>
                                        <button
                                            onClick={handleLogout}
                                            className="block w-full text-left px-4 py-2 text-red-600 hover:bg-gray-100"
                                        >
                                            Log Out
                                        </button>
                                    </div>
                                )}
                            </div>
                        ) : (
                            <>
                                <Link to="/login" className="text-gray-700 hover:text-blue-600">
                                    Login
                                </Link>
                                <Link to="/signup" className="bg-blue-600 text-white px-5 py-2 rounded-lg">
                                    Sign Up
                                </Link>
                            </>
                        )}
                    </nav>
                </div>
            </header>

            {/* Main */}
            <main className="flex-1 max-w-7xl mx-auto w-full px-6 py-10">

                <div className="flex flex-col md:flex-row justify-between items-center mb-10">
                    <h2 className="text-4xl font-bold mb-4 md:mb-0">
                        Our Products
                    </h2>

                    <select
                        value={sortOption}
                        onChange={(e) => setSortOption(e.target.value)}
                        className="p-2 border rounded-lg"
                    >
                        <option value="default">Featured</option>
                        <option value="price-asc">Price: Low to High</option>
                        <option value="price-desc">Price: High to Low</option>
                        <option value="rating-desc">Highest Rated</option>
                        <option value="name-asc">Name: A to Z</option>
                    </select>
                </div>

                {loading ? (
                    <p className="text-center">Loading products...</p>
                ) : (
                    <div className="grid gap-8 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
                        {products.map((product) => (
                            <div key={product.id} className="bg-white rounded-2xl shadow-md p-6 flex flex-col">

                                <img src={product.image_url} className="h-56 object-contain mb-4" />

                                <h3 className="text-xl font-semibold">{product.name}</h3>

                                <p>⭐ {product.rating} ({product.review_count})</p>

                                <p className="font-bold">${product.price}</p>

                                <p className={product.quantity === 0 ? "text-red-500" : "text-green-600"}>
                                    {product.quantity === 0 ? "Out of Stock" : `${product.quantity} left`}
                                </p>

                                <button
                                    onClick={() => addToCart(product)}
                                    disabled={product.quantity === 0}
                                    className="mt-auto bg-blue-600 text-white px-4 py-2 rounded-lg"
                                >
                                    Buy Now
                                </button>
                            </div>
                        ))}
                    </div>
                )}
            </main>
        </div>
    );
}