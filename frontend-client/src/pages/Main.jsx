import { Link, useNavigate } from 'react-router-dom';
import { useEffect, useState } from 'react';

export default function Main() {
    const API_URL = import.meta.env.VITE_API_URL;
    const navigate = useNavigate();

    const [products, setProducts]         = useState([]);
    const [loading, setLoading]           = useState(true);
    const [sortOption, setSortOption]     = useState("default");
    const [searchQuery, setSearchQuery]   = useState("");
    const [cartFeedback, setCartFeedback] = useState({});

    // Auth state
    const [isLoggedIn, setIsLoggedIn]     = useState(false);
    const [showDropdown, setShowDropdown] = useState(false);
    const [userData, setUserData]         = useState(null);

    useEffect(() => {
        const token = localStorage.getItem("token");
        const user  = localStorage.getItem("user");
        if (token) {
            setIsLoggedIn(true);
            if (user) setUserData(JSON.parse(user));
        }
    }, []);

    const handleLogout = () => {
        localStorage.removeItem("token");
        localStorage.removeItem("user");
        localStorage.removeItem("cart");
        setIsLoggedIn(false);
        setShowDropdown(false);
        setUserData(null);
    };

    // Fetch products with search + sort
    useEffect(() => {
        const fetchProducts = async () => {
            setLoading(true);
            try {
                let url = `${API_URL}/api/products?`;
                if (searchQuery) url += `search=${encodeURIComponent(searchQuery)}&`;
                if (sortOption !== "default") {
                    const sortMap = {
                        "price-asc":   "price_asc",
                        "price-desc":  "price_desc",
                        "rating-desc": "popular",
                        "name-asc":    "name_asc"
                    };
                    url += `sort=${sortMap[sortOption]}`;
                }
                const response = await fetch(url);
                const data     = await response.json();
                setProducts(Array.isArray(data) ? data : []);
            } catch (error) {
                console.error("Error fetching products:", error);
            } finally {
                setLoading(false);
            }
        };
        fetchProducts();
    }, [API_URL, searchQuery, sortOption]);

    // ── Add to Cart ───────────────────────────────────────────────────────────
    const addToCart = async (e, product) => {
        e.stopPropagation();
        if (product.quantity === 0) return;

        if (isLoggedIn) {
            // --- LOGGED-IN USER: Send to Backend ---
            try {
                const token = localStorage.getItem("token");
                const response = await fetch(`${API_URL}/api/cart/item`, {
                    method: 'PATCH',
                    headers: {
                        'Content-Type': 'application/json',
                        'Authorization': `Bearer ${token}` // Include the JWT!
                    },
                    body: JSON.stringify({
                        product_id: product.id,
                        quantity: 1 // Adding 1 item at a time
                    })
                });

                if (response.ok) {
                    setCartFeedback((prev) => ({ ...prev, [product.id]: "added" }));
                    setTimeout(() => setCartFeedback((prev) => ({ ...prev, [product.id]: null })), 1200);
                } else {
                    const errorData = await response.json();
                    console.error("Failed to add to backend cart:", errorData.error);
                    // Optional: You could set a specific feedback state here for errors
                }
            } catch (error) {
                console.error("Network error adding to cart:", error);
            }
        } else {
            // --- GUEST USER: Save to LocalStorage ---
            const cart = JSON.parse(localStorage.getItem("cart") || "[]");
            const idx  = cart.findIndex((item) => item.id === product.id);

            if (idx >= 0) {
                // Prevent adding more than what's in stock locally
                if (cart[idx].quantity >= product.quantity) {
                    setCartFeedback((prev) => ({ ...prev, [product.id]: "maxed" }));
                    setTimeout(() => setCartFeedback((prev) => ({ ...prev, [product.id]: null })), 1500);
                    return;
                }
                cart[idx].quantity = Math.min(cart[idx].quantity + 1, product.quantity);
            } else {
                cart.push({
                    id:        product.id,
                    name:      product.name,
                    price:     product.price,
                    image_url: product.image_url,
                    stock:     product.quantity,
                    quantity:  1,
                });
            }

            localStorage.setItem("cart", JSON.stringify(cart));
            setCartFeedback((prev) => ({ ...prev, [product.id]: "added" }));
            setTimeout(() => setCartFeedback((prev) => ({ ...prev, [product.id]: null })), 1200);
        }
    };

    return (
        <div className="min-h-screen bg-gray-50 flex flex-col">

            {/* Header */}
            <header className="w-full bg-white shadow-md relative z-50">
                <div className="max-w-7xl mx-auto px-4 sm:px-6 py-4 flex items-center gap-4">
                    <h1 className="text-xl sm:text-2xl font-bold text-gray-800 flex-shrink-0">
                        MyStore
                    </h1>
                    <div className="flex-1">
                        <input
                            type="text"
                            placeholder="Search products..."
                            value={searchQuery}
                            onChange={(e) => setSearchQuery(e.target.value)}
                            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
                        />
                    </div>
                    <nav className="flex items-center space-x-3 sm:space-x-6 flex-shrink-0">
                        <Link to="/shoppingcart" className="text-gray-700 hover:text-blue-600 transition text-sm sm:text-base">
                            🛒 <span className="hidden sm:inline">Cart</span>
                        </Link>
                        {isLoggedIn ? (
                            <div className="relative">
                                <button
                                    onClick={() => setShowDropdown(!showDropdown)}
                                    className="flex items-center space-x-1 text-gray-700 hover:text-blue-600 text-sm sm:text-base"
                                >
                                    <span className="font-medium">{userData?.name || "My Account"}</span>
                                    <span className="text-xs">▾</span>
                                </button>
                                {showDropdown && (
                                    <div className="absolute right-0 mt-3 w-48 bg-white border rounded-lg shadow-lg z-50">
                                        <Link to="/profile" className="block px-4 py-2 hover:bg-gray-100 text-sm">Profile</Link>
                                        <Link to="/orders"  className="block px-4 py-2 hover:bg-gray-100 text-sm">My Orders</Link>
                                        <button onClick={handleLogout} className="block w-full text-left px-4 py-2 text-red-600 hover:bg-gray-100 text-sm">
                                            Log Out
                                        </button>
                                    </div>
                                )}
                            </div>
                        ) : (
                            <>
                                <Link to="/login"  className="text-gray-700 hover:text-blue-600 text-sm sm:text-base">Login</Link>
                                <Link to="/signup" className="bg-blue-600 text-white px-3 sm:px-5 py-2 rounded-lg hover:bg-blue-700 text-sm sm:text-base">Sign Up</Link>
                            </>
                        )}
                    </nav>
                </div>
            </header>

            {/* Main */}
            <main className="flex-1 max-w-7xl mx-auto w-full px-4 sm:px-6 py-8 sm:py-10">
                <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center mb-8 gap-4">
                    <h2 className="text-3xl sm:text-4xl font-bold">Our Products</h2>
                    <select
                        value={sortOption}
                        onChange={(e) => setSortOption(e.target.value)}
                        className="p-2 border rounded-lg text-sm bg-white"
                    >
                        <option value="default">Featured</option>
                        <option value="price-asc">Price: Low to High</option>
                        <option value="price-desc">Price: High to Low</option>
                        <option value="rating-desc">Highest Rated</option>
                        <option value="name-asc">Name: A to Z</option>
                    </select>
                </div>

                {loading ? (
                    <p className="text-center text-gray-500 py-20">Loading products...</p>
                ) : products.length === 0 ? (
                    <p className="text-center text-gray-400 py-20">No products found.</p>
                ) : (
                    <div className="grid gap-6 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
                        {products.map((product) => {
                            const outOfStock = product.quantity === 0;
                            const feedback   = cartFeedback[product.id];

                            const discountedPrice = product.discount > 0
                                ? product.price * (1 - product.discount / 100)
                                : null;

                            return (
                                <div
                                    key={product.id}
                                    onClick={() => navigate(`/products/${product.id}`)}
                                    className="bg-white rounded-2xl shadow-md p-5 flex flex-col cursor-pointer hover:shadow-xl hover:-translate-y-1 transition-all duration-200"
                                >
                                    {/* Image */}
                                    <div className="w-full h-48 bg-gray-100 rounded-xl mb-4 flex items-center justify-center overflow-hidden">
                                        <img
                                            src={product.image_url}
                                            alt={product.name}
                                            className="h-full object-contain"
                                        />
                                    </div>

                                    {/* Category badge */}
                                    {product.category && (
                                        <span className="text-xs text-blue-600 font-semibold mb-1">
                                            {product.category}
                                        </span>
                                    )}

                                    {/* Name */}
                                    <h3 className="text-lg font-semibold text-gray-800 mb-1 line-clamp-2">
                                        {product.name}
                                    </h3>

                                    {/* Rating */}
                                    <p className="text-sm text-gray-500 mb-1">
                                        ⭐ {product.rating} ({product.review_count} reviews)
                                    </p>

                                    {/* Price */}
                                    <div className="mb-1">
                                        {discountedPrice ? (
                                            <div className="flex items-center gap-2">
                                                <p className="text-lg font-bold text-blue-600">
                                                    ${discountedPrice.toFixed(2)}
                                                </p>
                                                <p className="text-sm text-gray-400 line-through">
                                                    ${Number(product.price).toFixed(2)}
                                                </p>
                                                <span className="text-xs text-white bg-red-500 px-1.5 py-0.5 rounded">
                                                    -{product.discount}%
                                                </span>
                                            </div>
                                        ) : (
                                            <p className="text-lg font-bold text-gray-800">
                                                ${Number(product.price).toFixed(2)}
                                            </p>
                                        )}
                                    </div>

                                    {/* Stock */}
                                    <p className={`text-sm mb-4 ${outOfStock ? "text-red-500" : "text-green-600"}`}>
                                        {outOfStock ? "Out of Stock" : `${product.quantity} left in stock`}
                                    </p>

                                    {/* Add to Cart */}
                                    <button
                                        onClick={(e) => addToCart(e, product)}
                                        disabled={outOfStock}
                                        className={`mt-auto w-full py-2 px-4 rounded-lg font-semibold transition-colors duration-150 ${
                                            outOfStock
                                                ? "bg-gray-200 text-gray-400 cursor-not-allowed"
                                                : feedback === "added"
                                                ? "bg-green-500 text-white"
                                                : feedback === "maxed"
                                                ? "bg-red-400 text-white"
                                                : "bg-blue-600 text-white hover:bg-blue-700"
                                        }`}
                                    >
                                        {outOfStock
                                            ? "Out of Stock"
                                            : feedback === "added"
                                            ? "✓ Added!"
                                            : feedback === "maxed"
                                            ? "Max reached"
                                            : "Add to Cart"}
                                    </button>
                                </div>
                            );
                        })}
                    </div>
                )}
            </main>
        </div>
    );
}