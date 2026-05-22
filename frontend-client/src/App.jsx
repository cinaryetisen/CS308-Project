import { BrowserRouter, Routes, Route } from 'react-router-dom';
import Main from './pages/Main';
import Login from './pages/Login';
import Signup from './pages/Signup';
import ShoppingCart from './pages/ShoppingCart';
import ProductDetail from './pages/ProductDetail';
import Profile from './pages/Profile';
import MyOrders from "./pages/MyOrders";
import Payment from './pages/Payment';
import PMLayout from "./pages/product-manager/PMLayout";
import DeliveryManager from "./pages/product-manager/DeliveryManager";
import ReviewManager from "./pages/product-manager/ReviewManager";
import ProductManager from "./pages/product-manager/ProductManager";
import CategoryManager from "./pages/product-manager/CategoryManager";
import SMLayout from "./pages/sales-manager/SMLayout";
import PricingManager from "./pages/sales-manager/PricingManager";
import RevenueManager from "./pages/sales-manager/RevenueManager";
import MainLayout from './layouts/MainLayout';
import Invoice from './pages/Invoice';
import RefundManager from "./pages/sales-manager/RefundManager";
import Wishlist from './pages/Wishlist';
import InvoiceList from "./pages/sales-manager/InvoiceList";


function App() {
  return (
    <BrowserRouter>
      <Routes>
        
        {/* 2. Wrap the store pages inside the MainLayout */}
        <Route element={<MainLayout />}>
            <Route path="/" element={<Main />} />
            <Route path="/shoppingcart" element={<ShoppingCart />} />
            <Route path="/products/:id" element={<ProductDetail />} />
            <Route path="/profile" element={<Profile />} />
            <Route path="/orders" element={<MyOrders />} />
            <Route path="/payment" element={<Payment />} />
            <Route path="/invoice/:orderId" element={<Invoice />}/>
            <Route path="/wishlist" element={<Wishlist />} />
        </Route>

        {/* 3. Keep pages that SHOULDN'T have the header outside the wrapper */}
        <Route path="/login" element={<Login />} />
        <Route path="/signup" element={<Signup />} />
        
        {/* Product Manager Layout is completely separate */}
        <Route path="/pm" element={<PMLayout />}>
          <Route path="deliveries" element={<DeliveryManager />} />
          <Route path="reviews" element={<ReviewManager />} />
          <Route path="products" element={<ProductManager />} />
          <Route path="categories" element={<CategoryManager />} />
        </Route>
        <Route path="/sm" element={<SMLayout />}>
          <Route path="pricing" element={<PricingManager />} />
          <Route path="revenue" element={<RevenueManager />} />
          <Route path="invoices" element={<InvoiceList />} />
          <Route path="refunds" element={<RefundManager />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}


export default App;