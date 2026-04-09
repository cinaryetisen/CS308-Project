import { Link } from 'react-router-dom';

export default function ShoppingCart() {
    const cartItems = [];

    // Calculate Cost
    const totalCost = cartItems.reduce((sum, item) => sum + (item.price * item.quantity), 0);

    return (
        <div className="min-h-screen bg-gray-50 flex flex-col">
            
            {/* Header */}
            <header className="w-full bg-white shadow-md">
                <div className="max-w-7xl mx-auto px-6 py-4 flex justify-between items-center">
                    <h1 className="text-2xl font-bold text-gray-800">
                        MyStore
                    </h1>
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

            {/* Cart Content Area */}
            <main className="flex-1 max-w-4xl mx-auto w-full px-6 py-10">
                <h2 className="text-3xl font-bold mb-8 text-gray-800 border-b pb-4">
                    Your Shopping Cart
                </h2>

                {/* Logic: If cart is empty, show text. If full, show item boxes. */}
                {cartItems.length === 0 ? (
                    <div className="text-center mt-20">
                        <p className="text-xl text-gray-500 mb-6">Your cart is currently empty.</p>
                    </div>
                ) : (
                    <div className="space-y-6">
                        
                        {/* Item Boxes - These pop up automatically for every item in the cart */}
                        {cartItems.map((item, index) => (
                            <div key={index} className="bg-white border border-gray-200 rounded-xl p-4 flex justify-between items-center shadow-sm">
                                <div>
                                    <h3 className="text-lg font-semibold text-gray-800">{item.name}</h3>
                                    <p className="text-gray-500">Qty: {item.quantity}</p>
                                </div>
                                <div className="text-lg font-bold text-gray-800">
                                    ${(item.price * item.quantity).toFixed(2)}
                                </div>
                            </div>
                        ))}

                        {/* Cost Calculation and Payment Button */}
                        <div className="mt-10 bg-white border border-gray-200 rounded-xl p-6 flex flex-col items-end shadow-sm">
                            <h3 className="text-xl text-gray-600 mb-2">Total</h3>
                            <p className="text-3xl font-bold text-gray-800 mb-6">${totalCost.toFixed(2)}</p>
                            
                            <Link to="/payment" className="bg-green-600 text-white px-8 py-3 rounded-lg hover:bg-green-700 transition shadow-md text-lg font-bold">
                                Proceed to Payment
                            </Link>
                        </div>
                    </div>
                )}
            </main>
        </div>
    );
}