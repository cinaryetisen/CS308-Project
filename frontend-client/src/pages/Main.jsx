import { Link } from 'react-router-dom';
import { useEffect, useState } from 'react';

export default function Main() {
    const API_URL = import.meta.env.VITE_API_URL;

    const [products, setProducts] = useState([]);
    const [loading, setLoading] = useState(true);
    
    const [sortOption, setSortOption] = useState("default");

    // Fetch products from backend
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


    const sortedProducts = [...products].sort((a, b) => {
        if (sortOption === "price-asc") return a.price - b.price;
        if (sortOption === "price-desc") return b.price - a.price;
        if (sortOption === "rating-desc") return b.rating - a.rating;
        if (sortOption === "name-asc") return a.name.localeCompare(b.name);
        return 0; // "default" does nothing
    });

    return (
        <div className="min-h-screen bg-gray-50 flex flex-col">

            {/* Header */}
            <header className="w-full bg-white shadow-md">
                <div className="max-w-7xl mx-auto px-6 py-4 flex justify-between items-center">
                    <h1 className="text-2xl font-bold text-gray-800">
                        MyStore
                    </h1>
                    <nav className="flex items-center space-x-6">
                        <Link to="/login" className="text-gray-700 hover:text-blue-600 transition">
                            Login
                        </Link>
                        <Link to="/signup" className="bg-blue-600 text-white px-5 py-2 rounded-lg hover:bg-blue-700 transition">
                            Sign Up
                        </Link>
                        <Link to="/shoppingcart" className="text-gray-700 hover:text-blue-600 transition">
                            Shopping Cart
                        </Link>
                    </nav>
                </div>
            </header>

            {/* Main Content */}
            <main className="flex-1 max-w-7xl mx-auto w-full px-6 py-10">

                {/* Header & Sort Controls Container */}
                <div className="flex flex-col md:flex-row justify-between items-center mb-10">
                    <h2 className="text-4xl font-bold text-center md:text-left mb-4 md:mb-0">
                        Our Products
                    </h2>

                    {/* NEW: Sorting Dropdown Menu */}
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

                {/* Loading */}
                {loading ? (
                    <p className="text-center text-gray-600">Loading products...</p>
                ) : (
                    <div className="grid gap-8 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
                        {/* Notice we are mapping over sortedProducts now, not products */}
                        {sortedProducts.map((product) => (
                            <div
                                key={product.id}
                                className="bg-white rounded-2xl shadow-md p-6 flex flex-col hover:shadow-xl transition duration-300"
                            >
                                {/* Image */}
                                <div className="w-full h-56 bg-gray-200 rounded-xl mb-5 flex items-center justify-center">
                                    <img
                                        src={product.image_url}
                                        alt={product.name}
                                        className="h-full object-contain"
                                    />
                                </div>

                                {/* Product Name */}
                                <h3 className="text-xl font-semibold mb-2">
                                    {product.name}
                                </h3>

                                {/* Rating */}
                                <div className="flex items-center mb-2">
                                    <span className="text-yellow-500 mr-2">
                                        {"★".repeat(Math.floor(product.rating || 0))}
                                    </span>
                                    <span className="text-gray-600 text-sm">
                                        {product.rating} ({product.review_count || 0})
                                    </span>
                                </div>

                                {/* Price */}
                                <p className="text-lg font-medium text-gray-800 mb-2">
                                    ${product.price}
                                </p>

                                {/* Stock Indicator */}
                                <p className={`text-sm mb-4 ${
                                    product.quantity === 0 ? "text-red-500" : "text-green-600"
                                }`}>
                                    {product.quantity === 0
                                        ? "Out of Stock"
                                        : `${product.quantity} left in stock`}
                                </p>

                                {/* Button */}
                                <button
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