import { useEffect, useState } from "react";
import { NavLink, Outlet, useNavigate } from "react-router-dom";

function getTokenPayload() {
    try {
        const token = localStorage.getItem("token");
        if (!token) return null;
        return JSON.parse(atob(token.split(".")[1]));
    } catch {
        return null;
    }
}

const NAV_LINKS = [
    { to: "/pm/deliveries", label: "📦 Deliveries" },
    // Add future PM pages here
];

export default function PMLayout() {
    const [checking, setChecking] = useState(true);
    const [authorized, setAuthorized] = useState(false);
    const navigate = useNavigate();

    useEffect(() => {
        const payload = getTokenPayload();
        if (!payload || payload.role !== "product_manager") {
            setAuthorized(false);
        } else {
            setAuthorized(true);
        }
        setChecking(false);
    }, []);

    if (checking) {
        return (
            <div className="min-h-screen bg-gray-100 flex items-center justify-center">
                <div className="w-8 h-8 border-4 border-gray-200 border-t-blue-600 rounded-full animate-spin" />
            </div>
        );
    }

    if (!authorized) {
        return (
            <div className="min-h-screen bg-gray-100 flex items-center justify-center px-4">
                <div className="bg-white border border-gray-200 rounded-xl p-10 text-center max-w-sm w-full">
                    <p className="text-4xl mb-4">🚫</p>
                    <h1 className="text-lg font-bold text-gray-900 mb-1">Access Denied</h1>
                    <p className="text-sm text-gray-500 mb-6">
                        This area is restricted to product managers only.
                    </p>
                    <button
                        className="bg-blue-600 hover:bg-blue-700 text-white text-sm font-semibold px-5 py-2 rounded-lg transition-colors"
                        onClick={() => navigate("/")}
                    >
                        Back to store
                    </button>
                </div>
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-gray-100 flex">

            {/* Sidebar */}
            <aside className="w-56 shrink-0 bg-white border-r border-gray-200 flex flex-col">
                <div className="px-5 py-5 border-b border-gray-100">
                    <p className="text-xs font-semibold text-gray-400 uppercase tracking-wide mb-0.5">
                        MyStore
                    </p>
                    <h2 className="text-base font-bold text-gray-900">Manager Panel</h2>
                </div>

                <nav className="flex-1 px-3 py-4 flex flex-col gap-1">
                    {NAV_LINKS.map(({ to, label }) => (
                        <NavLink
                            key={to}
                            to={to}
                            className={({ isActive }) =>
                                `text-sm font-medium px-3 py-2 rounded-lg transition-colors ${
                                    isActive
                                        ? "bg-blue-50 text-blue-600"
                                        : "text-gray-600 hover:bg-gray-100 hover:text-gray-900"
                                }`
                            }
                        >
                            {label}
                        </NavLink>
                    ))}
                </nav>

                <div className="px-3 py-4 border-t border-gray-100">
                    <button
                        className="w-full text-sm font-medium text-gray-500 hover:text-gray-900 px-3 py-2 rounded-lg hover:bg-gray-100 transition-colors text-left"
                        onClick={() => navigate("/")}
                    >
                        ← Back to store
                    </button>
                </div>
            </aside>

            {/* Main content */}
            <main className="flex-1 overflow-y-auto py-8 px-6">
                <Outlet />
            </main>

        </div>
    );
}