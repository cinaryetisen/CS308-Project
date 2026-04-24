import { BrowserRouter, Routes, Route } from 'react-router-dom';
import Main from './pages/Main';
import Login from './pages/Login';
import Signup from './pages/Signup';
import ShoppingCart from './pages/ShoppingCart';
import ProductDetail from './pages/ProductDetail';
import Profile from './pages/Profile';
import MyOrders from "./pages/MyOrders";
import PMLayout from "./pages/product-manager/PMLayout";
import DeliveryManager from "./pages/product-manager/DeliveryManager";


function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Main />} />
        <Route path="/login" element={<Login />} />
        <Route path="/signup" element={<Signup />} />
        <Route path="/shoppingcart" element={<ShoppingCart />} />
        <Route path="/products/:id" element={<ProductDetail />} />
        <Route path="/profile" element={<Profile />} />
        <Route path="/orders" element={<MyOrders />} />
        <Route path="/pm" element={<PMLayout />}>
          <Route path="deliveries" element={<DeliveryManager />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}

export default App;