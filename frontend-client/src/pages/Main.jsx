import { Link } from 'react-router-dom';

export default function Main() {
    const products = [
        { id: 1, name: "Product 1", price: "$10", rating: 4.5, stock: 12 },
        { id: 2, name: "Product 2", price: "$20", rating: 3.8, stock: 5 },
        { id: 3, name: "Product 3", price: "$30", rating: 4.9, stock: 0 },
        { id: 4, name: "Product 4", price: "$5", rating: 2.1, stock: 100 },
    ];

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
                    </nav>

                </div>
            </header>

            {/* Main Content */}
            <main className="flex-1 max-w-7xl mx-auto w-full px-6 py-10">

                <h2 className="text-4xl font-bold mb-10 text-center">
                    Our Products
                </h2>

                {/* Flexible Grid */}
                <div className="grid gap-8 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
                    {products.map((product) => (
                        <div
                            key={product.id}
                            className="bg-white rounded-2xl shadow-md p-6 flex flex-col hover:shadow-xl transition duration-300"
                        >
                            {/* Image */}
                            <div className="w-full h-56 bg-gray-200 rounded-xl mb-5 flex items-center justify-center">
                                <span className="text-gray-500">Image</span>
                            </div>

                            {/* Product Name */}
                            <h3 className="text-xl font-semibold mb-2">
                                {product.name}
                            </h3>

                            {/* Rating */}
                            <div className="flex items-center mb-2">
                <span className="text-yellow-500 mr-2">
                  {"★".repeat(Math.floor(product.rating))}
                </span>
                                <span className="text-gray-600 text-sm">
                  {product.rating}
                </span>
                            </div>

                            {/* Price */}
                            <p className="text-lg font-medium text-gray-800 mb-2">
                                {product.price}
                            </p>

                            {/* Stock Indicator */}
                            <p className={`text-sm mb-4 ${product.stock === 0 ? "text-red-500" : "text-green-600"}`}>
                                {product.stock === 0
                                    ? "Out of Stock"
                                    : `${product.stock} left in stock`}
                            </p>

                            {/* Button */}
                            <button
                                disabled={product.stock === 0}
                                className={`mt-auto px-4 py-2 rounded-lg transition ${
                                    product.stock === 0
                                        ? "bg-gray-300 cursor-not-allowed"
                                        : "bg-blue-600 text-white hover:bg-blue-700"
                                }`}
                            >
                                {product.stock === 0 ? "Unavailable" : "Buy Now"}
                            </button>
                        </div>
                    ))}
                </div>

            </main>

        </div>
    );
}