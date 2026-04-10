import { useEffect, useState } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';

export default function ProductDetail() {
    const { id }   = useParams();
    const navigate = useNavigate();
    const API_URL  = import.meta.env.VITE_API_URL;

    const [product, setProduct]   = useState(null);
    const [loading, setLoading]   = useState(true);
    const [error, setError]       = useState("");
    const [quantity, setQuantity] = useState(1);
    const [cartMsg, setCartMsg]   = useState(null);

    // Auth
    const [isLoggedIn, setIsLoggedIn]     = useState(false);
    const [userData, setUserData]         = useState(null);
    const [showDropdown, setShowDropdown] = useState(false);

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

    // Fetch single product
    useEffect(() => {
        const fetchProduct = async () => {
            try {
                const res  = await fetch(`${API_URL}/api/products/${id}`);
                const text = await res.text();

                let data;
                try {
                    data = JSON.parse(text);
                } catch {
                    setError("Unexpected response from server.");
                    return;
                }

                if (!res.ok) {
                    setError(data.error || "Product not found.");
                    return;
                }

                setProduct(data);
            } catch (err) {
                console.error(err);
                setError("Server error. Please try again later.");
            } finally {
                setLoading(false);
            }
        };
        fetchProduct();
    }, [id]);

    // Add to Cart
    const handleAddToCart = () => {
        if (!product || product.quantity === 0) return;

        const cart = JSON.parse(localStorage.getItem("cart") || "[]");
        const idx  = cart.findIndex((item) => item.id === product.id);

        if (idx >= 0) {
            if (cart[idx].quantity >= product.quantity) {
                setCartMsg("maxed");
                setTimeout(() => setCartMsg(null), 1500);
                return;
            }
            cart[idx].quantity = Math.min(cart[idx].quantity + quantity, product.quantity);
        } else {
            cart.push({
                id:        product.id,
                name:      product.name,
                price:     product.price,
                image_url: product.image_url,
                stock:     product.quantity,
                quantity:  Math.min(quantity, product.quantity),
            });
        }

        localStorage.setItem("cart", JSON.stringify(cart));
        setCartMsg("added");
        setTimeout(() => setCartMsg(null), 1500);
    };

    // Loading
    if (loading) {
        return (
            <div className="flex justify-center items-center min-h-screen">
                <p className="text-gray-500">Loading product...</p>
            </div>
        );
    }

    // Error
    if (error || !product) {
        return (
            <div className="flex flex-col items-center justify-center min-h-screen gap-4">
                <p className="text-red-500 text-lg">{error || "Product not found."}</p>
                <button
                    onClick={() => navigate("/")}
                    className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
                >
                    Back to Products
                </button>
            </div>
        );
    }

    const outOfStock      = product.quantity === 0;
    const discountedPrice = product.discount > 0
        ? product.price * (1 - product.discount / 100)
        : null;

    return (
        <div className="min-h-screen bg-gray-50 flex flex-col">

            {/* Header */}
            <header className="w-full bg-white shadow-md relative z-50">
                <div className="max-w-7xl mx-auto px-4 sm:px-6 py-4 flex items-center gap-4">
                    <Link to="/" className="text-xl sm:text-2xl font-bold text-gray-800 flex-shrink-0">
                        MyStore
                    </Link>
                    <div className="flex-1" />
                    <nav className="flex items-center space-x-3 sm:space-x-6 flex-shrink-0">
                        <Link to="/shoppingcart" className="text-gray-700 hover:text-blue-600 text-sm sm:text-base">
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

            {/* Breadcrumb */}
            <div className="max-w-7xl mx-auto w-full px-4 sm:px-6 pt-5 text-sm text-gray-400 flex gap-2">
                <Link to="/" className="hover:text-blue-600">Home</Link>
                <span>/</span>
                <span className="text-gray-700 font-medium truncate">{product.name}</span>
            </div>

            {/* Product Detail */}
            <main className="flex-1 max-w-7xl mx-auto w-full px-4 sm:px-6 py-8">
                <div className="bg-white rounded-2xl shadow-md overflow-hidden">
                    <div className="flex flex-col md:flex-row">

                        {/* LEFT: Image */}
                        <div className="md:w-1/2 bg-gray-100 flex items-center justify-center p-8 min-h-72 md:min-h-96">
                            {product.image_url ? (
                                <img
                                    src={product.image_url}
                                    alt={product.name}
                                    className="max-h-80 w-full object-contain"
                                />
                            ) : (
                                <div className="text-gray-300 text-8xl">📦</div>
                            )}
                        </div>

                        {/* RIGHT: Info */}
                        <div className="md:w-1/2 p-6 sm:p-8 flex flex-col gap-4">

                            {product.category && (
                                <span className="text-xs text-blue-600 font-semibold uppercase tracking-wide">
                                    {product.category}
                                </span>
                            )}

                            <h1 className="text-2xl sm:text-3xl font-bold text-gray-900 leading-snug">
                                {product.name}
                            </h1>

                            <div className="flex items-center gap-2">
                                <span className="text-yellow-400 text-lg">
                                    {"★".repeat(Math.floor(product.rating))}
                                    {"☆".repeat(5 - Math.floor(product.rating))}
                                </span>
                                <span className="text-gray-500 text-sm">
                                    {product.rating} ({product.review_count} reviews)
                                </span>
                            </div>

                            <div>
                                {discountedPrice ? (
                                    <div className="flex items-center gap-3">
                                        <p className="text-3xl font-bold text-blue-600">${discountedPrice.toFixed(2)}</p>
                                        <p className="text-lg text-gray-400 line-through">${Number(product.price).toFixed(2)}</p>
                                        <span className="text-sm text-white bg-red-500 px-2 py-0.5 rounded-full">-{product.discount}%</span>
                                    </div>
                                ) : (
                                    <p className="text-3xl font-bold text-blue-600">
                                        ${Number(product.price).toFixed(2)}
                                    </p>
                                )}
                            </div>

                            <p className={`text-sm font-medium ${outOfStock ? "text-red-500" : "text-green-600"}`}>
                                {outOfStock ? "Out of Stock" : `${product.quantity} left in stock`}
                            </p>

                            {product.description && (
                                <p className="text-gray-600 text-sm leading-relaxed border-t pt-4">
                                    {product.description}
                                </p>
                            )}

                            <div className="text-xs text-gray-400 flex flex-col gap-1 border-t pt-3">
                                {product.model         && <span>Model: {product.model}</span>}
                                {product.serial_number && <span>Serial No: {product.serial_number}</span>}
                                {product.warranty      && <span>Warranty: {product.warranty}</span>}
                                {product.distributor   && <span>Distributor: {product.distributor}</span>}
                            </div>

                            {product.tags && product.tags.length > 0 && (
                                <div className="flex flex-wrap gap-2">
                                    {product.tags.map((tag) => (
                                        <span key={tag} className="text-xs bg-gray-100 text-gray-600 px-2 py-1 rounded-full">
                                            {tag}
                                        </span>
                                    ))}
                                </div>
                            )}

                            {cartMsg === "added" && (
                                <div className="px-4 py-2 text-sm text-green-700 bg-green-100 border border-green-300 rounded-lg">
                                    ✓ Added to cart!
                                </div>
                            )}
                            {cartMsg === "maxed" && (
                                <div className="px-4 py-2 text-sm text-red-700 bg-red-100 border border-red-300 rounded-lg">
                                    Maximum stock reached.
                                </div>
                            )}

                            {!outOfStock && (
                                <div className="flex items-center gap-3">
                                    <span className="text-sm font-medium text-gray-700">Quantity:</span>
                                    <div className="flex items-center border border-gray-300 rounded-lg overflow-hidden">
                                        <button
                                            onClick={() => setQuantity((q) => Math.max(1, q - 1))}
                                            className="px-3 py-2 bg-gray-100 hover:bg-gray-200 font-bold text-gray-700 transition"
                                        >−</button>
                                        <span className="px-4 py-2 text-sm font-semibold min-w-[2.5rem] text-center">
                                            {quantity}
                                        </span>
                                        <button
                                            onClick={() => setQuantity((q) => Math.min(product.quantity, q + 1))}
                                            className="px-3 py-2 bg-gray-100 hover:bg-gray-200 font-bold text-gray-700 transition"
                                        >+</button>
                                    </div>
                                </div>
                            )}

                            <button
                                onClick={handleAddToCart}
                                disabled={outOfStock}
                                className={`w-full py-3 rounded-lg font-bold text-white transition ${
                                    outOfStock
                                        ? "bg-gray-200 text-gray-400 cursor-not-allowed"
                                        : "bg-blue-600 hover:bg-blue-700"
                                }`}
                            >
                                {outOfStock ? "Out of Stock" : "Add to Cart"}
                            </button>

                            <Link to="/" className="text-center text-sm text-gray-400 hover:text-blue-600 transition">
                                ← Back to Products
                            </Link>
                        </div>
                    </div>
                </div>
            </main>
        </div>
    );
}