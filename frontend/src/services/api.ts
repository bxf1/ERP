import {
  mockSuppliers, mockCustomers, mockPurchaseOrders,
  mockSalesOrders, mockInventory, mockStocktaking, dashboardData,
  type Supplier, type Customer, type PurchaseOrder, type SalesOrder,
  type InventoryRecord, type Stocktaking,
} from '../mock/data';

const delay = (ms: number = 300) => new Promise(resolve => setTimeout(resolve, ms));

// Supplier API
export async function getSuppliers(params?: { page?: number; pageSize?: number; keyword?: string }) {
  await delay();
  let data = [...mockSuppliers];
  if (params?.keyword) {
    const kw = params.keyword.toLowerCase();
    data = data.filter(s => s.name.includes(kw) || s.code.toLowerCase().includes(kw) || s.contact.includes(kw));
  }
  const page = params?.page || 1;
  const pageSize = params?.pageSize || 10;
  return { data: data.slice((page - 1) * pageSize, page * pageSize), total: data.length, success: true };
}

export async function getSupplier(id: string) {
  await delay();
  const item = mockSuppliers.find(s => s.id === id);
  return { data: item, success: !!item };
}

export async function createSupplier(data: Partial<Supplier>) {
  await delay();
  const newItem: Supplier = { id: `sup-${Date.now()}`, code: '', name: '', contact: '', phone: '', email: '', address: '', bankAccount: '', taxId: '', status: 'active', createdAt: new Date().toISOString().slice(0, 10), ...data };
  mockSuppliers.unshift(newItem);
  return { data: newItem, success: true };
}

export async function updateSupplier(id: string, data: Partial<Supplier>) {
  await delay();
  const idx = mockSuppliers.findIndex(s => s.id === id);
  if (idx > -1) { mockSuppliers[idx] = { ...mockSuppliers[idx], ...data }; return { data: mockSuppliers[idx], success: true }; }
  return { success: false, message: 'Not found' };
}

// Customer API
export async function getCustomers(params?: { page?: number; pageSize?: number; keyword?: string }) {
  await delay();
  let data = [...mockCustomers];
  if (params?.keyword) {
    const kw = params.keyword.toLowerCase();
    data = data.filter(c => c.name.includes(kw) || c.code.toLowerCase().includes(kw) || c.contact.includes(kw));
  }
  const page = params?.page || 1;
  const pageSize = params?.pageSize || 10;
  return { data: data.slice((page - 1) * pageSize, page * pageSize), total: data.length, success: true };
}

export async function getCustomer(id: string) {
  await delay();
  const item = mockCustomers.find(c => c.id === id);
  return { data: item, success: !!item };
}

export async function createCustomer(data: Partial<Customer>) {
  await delay();
  const newItem: Customer = { id: `cust-${Date.now()}`, code: '', name: '', contact: '', phone: '', email: '', address: '', creditLimit: 0, status: 'active', createdAt: new Date().toISOString().slice(0, 10), ...data };
  mockCustomers.unshift(newItem);
  return { data: newItem, success: true };
}

export async function updateCustomer(id: string, data: Partial<Customer>) {
  await delay();
  const idx = mockCustomers.findIndex(c => c.id === id);
  if (idx > -1) { mockCustomers[idx] = { ...mockCustomers[idx], ...data }; return { data: mockCustomers[idx], success: true }; }
  return { success: false, message: 'Not found' };
}

// Purchase Order API
export async function getPurchaseOrders(params?: { page?: number; pageSize?: number; status?: string }) {
  await delay();
  let data = [...mockPurchaseOrders];
  if (params?.status) data = data.filter(o => o.status === params.status);
  const page = params?.page || 1;
  const pageSize = params?.pageSize || 10;
  return { data: data.slice((page - 1) * pageSize, page * pageSize), total: data.length, success: true };
}

export async function getPurchaseOrder(id: string) {
  await delay();
  const item = mockPurchaseOrders.find(o => o.id === id);
  return { data: item, success: !!item };
}

export async function createPurchaseOrder(data: Partial<PurchaseOrder>) {
  await delay();
  const newItem: PurchaseOrder = {
    id: `po-${Date.now()}`, orderNo: '', supplierId: '', supplierName: '', orderDate: '', deliveryDate: '',
    items: [], totalAmount: 0, status: 'draft', remark: '', createdAt: new Date().toISOString().slice(0, 10), ...data,
  };
  mockPurchaseOrders.unshift(newItem);
  return { data: newItem, success: true };
}

export async function updatePurchaseOrder(id: string, data: Partial<PurchaseOrder>) {
  await delay();
  const idx = mockPurchaseOrders.findIndex(o => o.id === id);
  if (idx > -1) { mockPurchaseOrders[idx] = { ...mockPurchaseOrders[idx], ...data }; return { data: mockPurchaseOrders[idx], success: true }; }
  return { success: false, message: 'Not found' };
}

// Sales Order API
export async function getSalesOrders(params?: { page?: number; pageSize?: number; status?: string }) {
  await delay();
  let data = [...mockSalesOrders];
  if (params?.status) data = data.filter(o => o.status === params.status);
  const page = params?.page || 1;
  const pageSize = params?.pageSize || 10;
  return { data: data.slice((page - 1) * pageSize, page * pageSize), total: data.length, success: true };
}

export async function getSalesOrder(id: string) {
  await delay();
  const item = mockSalesOrders.find(o => o.id === id);
  return { data: item, success: !!item };
}

export async function createSalesOrder(data: Partial<SalesOrder>) {
  await delay();
  const newItem: SalesOrder = {
    id: `so-${Date.now()}`, orderNo: '', customerId: '', customerName: '', orderDate: '', deliveryDate: '',
    items: [], totalAmount: 0, status: 'draft', remark: '', createdAt: new Date().toISOString().slice(0, 10), ...data,
  };
  mockSalesOrders.unshift(newItem);
  return { data: newItem, success: true };
}

export async function updateSalesOrder(id: string, data: Partial<SalesOrder>) {
  await delay();
  const idx = mockSalesOrders.findIndex(o => o.id === id);
  if (idx > -1) { mockSalesOrders[idx] = { ...mockSalesOrders[idx], ...data }; return { data: mockSalesOrders[idx], success: true }; }
  return { success: false, message: 'Not found' };
}

// Inventory API
export async function getInventory(params?: { page?: number; pageSize?: number; keyword?: string }) {
  await delay();
  let data = [...mockInventory];
  if (params?.keyword) {
    const kw = params.keyword.toLowerCase();
    data = data.filter(r => r.productName.includes(kw) || r.productCode.toLowerCase().includes(kw));
  }
  const page = params?.page || 1;
  const pageSize = params?.pageSize || 10;
  return { data: data.slice((page - 1) * pageSize, page * pageSize), total: data.length, success: true };
}

// Stocktaking API
export async function getStocktakingList(params?: { page?: number; pageSize?: number }) {
  await delay();
  const page = params?.page || 1;
  const pageSize = params?.pageSize || 10;
  return { data: mockStocktaking.slice((page - 1) * pageSize, page * pageSize), total: mockStocktaking.length, success: true };
}

export async function getStocktaking(id: string) {
  await delay();
  const item = mockStocktaking.find(s => s.id === id);
  return { data: item, success: !!item };
}

export async function createStocktaking(data: Partial<Stocktaking>) {
  await delay();
  const newItem: Stocktaking = {
    id: `st-${Date.now()}`, taskNo: '', warehouseId: '', warehouseName: '',
    startDate: '', endDate: '', status: 'pending', items: [], createdAt: new Date().toISOString().slice(0, 10), ...data,
  };
  mockStocktaking.unshift(newItem);
  return { data: newItem, success: true };
}

// Dashboard API
export async function getDashboardData() {
  await delay();
  return { data: dashboardData, success: true };
}
