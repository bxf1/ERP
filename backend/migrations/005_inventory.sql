-- Migration 005: Inventory / Purchase / Sales / Stocktaking tables
-- Adds install_tenant_v4_tables for business data tables under each tenant schema.

CREATE OR REPLACE FUNCTION install_tenant_v4_tables(_schema text)
RETURNS void AS $$
BEGIN
    -- =========================================================================
    -- SUPPLIERS
    -- =========================================================================
    EXECUTE format($f$
        CREATE TABLE IF NOT EXISTS %I.suppliers (
            id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            code         VARCHAR(100) NOT NULL,
            name         VARCHAR(300) NOT NULL,
            contact      VARCHAR(100) NOT NULL DEFAULT '',
            phone        VARCHAR(50)  NOT NULL DEFAULT '',
            email        VARCHAR(200) NOT NULL DEFAULT '',
            address      TEXT         NOT NULL DEFAULT '',
            bank_account VARCHAR(100) NOT NULL DEFAULT '',
            tax_id       VARCHAR(100) NOT NULL DEFAULT '',
            status       VARCHAR(20)  NOT NULL DEFAULT 'active'
                         CHECK (status IN ('active', 'inactive')),
            created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
            updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now()
        );
        CREATE INDEX IF NOT EXISTS idx_suppliers_code ON %I.suppliers (code);
        CREATE INDEX IF NOT EXISTS idx_suppliers_status ON %I.suppliers (status);
    $f$, _schema, _schema, _schema);

    -- =========================================================================
    -- CUSTOMERS
    -- =========================================================================
    EXECUTE format($f$
        CREATE TABLE IF NOT EXISTS %I.customers (
            id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            code         VARCHAR(100) NOT NULL,
            name         VARCHAR(300) NOT NULL,
            contact      VARCHAR(100) NOT NULL DEFAULT '',
            phone        VARCHAR(50)  NOT NULL DEFAULT '',
            email        VARCHAR(200) NOT NULL DEFAULT '',
            address      TEXT         NOT NULL DEFAULT '',
            credit_limit DECIMAL(18,2) NOT NULL DEFAULT 0,
            status       VARCHAR(20)  NOT NULL DEFAULT 'active'
                         CHECK (status IN ('active', 'inactive')),
            created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
            updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now()
        );
        CREATE INDEX IF NOT EXISTS idx_customers_code ON %I.customers (code);
        CREATE INDEX IF NOT EXISTS idx_customers_status ON %I.customers (status);
    $f$, _schema, _schema, _schema);

    -- =========================================================================
    -- PURCHASE ORDERS
    -- =========================================================================
    EXECUTE format($f$
        CREATE TABLE IF NOT EXISTS %I.purchase_orders (
            id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            order_no      VARCHAR(100) NOT NULL,
            supplier_id   VARCHAR(100) NOT NULL DEFAULT '',
            supplier_name VARCHAR(300) NOT NULL DEFAULT '',
            order_date    VARCHAR(20)  NOT NULL DEFAULT '',
            delivery_date VARCHAR(20)  NOT NULL DEFAULT '',
            total_amount  DECIMAL(18,2) NOT NULL DEFAULT 0,
            status        VARCHAR(20)  NOT NULL DEFAULT 'draft'
                          CHECK (status IN ('draft','submitted','approved','received','cancelled')),
            remark        TEXT         NOT NULL DEFAULT '',
            created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
            updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now()
        );
        CREATE INDEX IF NOT EXISTS idx_po_order_no ON %I.purchase_orders (order_no);
        CREATE INDEX IF NOT EXISTS idx_po_status ON %I.purchase_orders (status);

        CREATE TABLE IF NOT EXISTS %I.purchase_order_items (
            id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            purchase_order_id UUID NOT NULL REFERENCES %I.purchase_orders(id) ON DELETE CASCADE,
            product_id        VARCHAR(100) NOT NULL DEFAULT '',
            product_code      VARCHAR(100) NOT NULL DEFAULT '',
            product_name      VARCHAR(300) NOT NULL DEFAULT '',
            specification     VARCHAR(200) NOT NULL DEFAULT '',
            unit              VARCHAR(50)  NOT NULL DEFAULT '',
            quantity          DECIMAL(18,4) NOT NULL DEFAULT 0,
            unit_price        DECIMAL(18,4) NOT NULL DEFAULT 0,
            amount            DECIMAL(18,2) NOT NULL DEFAULT 0,
            created_at        TIMESTAMPTZ  NOT NULL DEFAULT now()
        );
        CREATE INDEX IF NOT EXISTS idx_poi_order ON %I.purchase_order_items (purchase_order_id);
    $f$, _schema, _schema, _schema, _schema, _schema);

    -- =========================================================================
    -- SALES ORDERS
    -- =========================================================================
    EXECUTE format($f$
        CREATE TABLE IF NOT EXISTS %I.sales_orders (
            id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            order_no      VARCHAR(100) NOT NULL,
            customer_id   VARCHAR(100) NOT NULL DEFAULT '',
            customer_name VARCHAR(300) NOT NULL DEFAULT '',
            order_date    VARCHAR(20)  NOT NULL DEFAULT '',
            delivery_date VARCHAR(20)  NOT NULL DEFAULT '',
            total_amount  DECIMAL(18,2) NOT NULL DEFAULT 0,
            status        VARCHAR(20)  NOT NULL DEFAULT 'draft'
                          CHECK (status IN ('draft','confirmed','shipped','invoiced','cancelled')),
            remark        TEXT         NOT NULL DEFAULT '',
            created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
            updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now()
        );
        CREATE INDEX IF NOT EXISTS idx_so_order_no ON %I.sales_orders (order_no);
        CREATE INDEX IF NOT EXISTS idx_so_status ON %I.sales_orders (status);

        CREATE TABLE IF NOT EXISTS %I.sales_order_items (
            id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            sales_order_id  UUID NOT NULL REFERENCES %I.sales_orders(id) ON DELETE CASCADE,
            product_id      VARCHAR(100) NOT NULL DEFAULT '',
            product_code    VARCHAR(100) NOT NULL DEFAULT '',
            product_name    VARCHAR(300) NOT NULL DEFAULT '',
            specification   VARCHAR(200) NOT NULL DEFAULT '',
            unit            VARCHAR(50)  NOT NULL DEFAULT '',
            quantity        DECIMAL(18,4) NOT NULL DEFAULT 0,
            unit_price      DECIMAL(18,4) NOT NULL DEFAULT 0,
            amount          DECIMAL(18,2) NOT NULL DEFAULT 0,
            created_at      TIMESTAMPTZ  NOT NULL DEFAULT now()
        );
        CREATE INDEX IF NOT EXISTS idx_soi_order ON %I.sales_order_items (sales_order_id);
    $f$, _schema, _schema, _schema, _schema, _schema);

    -- =========================================================================
    -- INVENTORY
    -- =========================================================================
    EXECUTE format($f$
        CREATE TABLE IF NOT EXISTS %I.inventory (
            id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            product_id     VARCHAR(100) NOT NULL DEFAULT '',
            product_code   VARCHAR(100) NOT NULL DEFAULT '',
            product_name   VARCHAR(300) NOT NULL DEFAULT '',
            specification  VARCHAR(200) NOT NULL DEFAULT '',
            unit           VARCHAR(50)  NOT NULL DEFAULT '',
            warehouse_id   VARCHAR(100) NOT NULL DEFAULT '',
            warehouse_name VARCHAR(200) NOT NULL DEFAULT '',
            quantity       DECIMAL(18,4) NOT NULL DEFAULT 0,
            safety_stock   DECIMAL(18,4) NOT NULL DEFAULT 0,
            cost_price     DECIMAL(18,4) NOT NULL DEFAULT 0,
            created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
            updated_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
            UNIQUE (product_id, warehouse_id)
        );
        CREATE INDEX IF NOT EXISTS idx_inventory_product ON %I.inventory (product_code);
    $f$, _schema, _schema);

    -- =========================================================================
    -- STOCKTAKING
    -- =========================================================================
    EXECUTE format($f$
        CREATE TABLE IF NOT EXISTS %I.stocktaking (
            id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            task_no        VARCHAR(100) NOT NULL,
            warehouse_id   VARCHAR(100) NOT NULL DEFAULT '',
            warehouse_name VARCHAR(200) NOT NULL DEFAULT '',
            start_date     VARCHAR(20)  NOT NULL DEFAULT '',
            end_date       VARCHAR(20)  NOT NULL DEFAULT '',
            status         VARCHAR(20)  NOT NULL DEFAULT 'pending'
                           CHECK (status IN ('pending','in_progress','completed','cancelled')),
            remark         TEXT         NOT NULL DEFAULT '',
            created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
            updated_at     TIMESTAMPTZ  NOT NULL DEFAULT now()
        );
        CREATE INDEX IF NOT EXISTS idx_st_task_no ON %I.stocktaking (task_no);

        CREATE TABLE IF NOT EXISTS %I.stocktaking_items (
            id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            stocktaking_id  UUID NOT NULL REFERENCES %I.stocktaking(id) ON DELETE CASCADE,
            product_id      VARCHAR(100) NOT NULL DEFAULT '',
            product_code    VARCHAR(100) NOT NULL DEFAULT '',
            product_name    VARCHAR(300) NOT NULL DEFAULT '',
            specification   VARCHAR(200) NOT NULL DEFAULT '',
            unit            VARCHAR(50)  NOT NULL DEFAULT '',
            book_quantity   DECIMAL(18,4) NOT NULL DEFAULT 0,
            actual_quantity DECIMAL(18,4) NOT NULL DEFAULT 0,
            diff_quantity   DECIMAL(18,4) NOT NULL DEFAULT 0,
            remark          TEXT NOT NULL DEFAULT '',
            created_at      TIMESTAMPTZ  NOT NULL DEFAULT now()
        );
        CREATE INDEX IF NOT EXISTS idx_sti_task ON %I.stocktaking_items (stocktaking_id);
    $f$, _schema, _schema, _schema, _schema, _schema);

    INSERT INTO public.schema_migrations (schema_name, version, name)
    VALUES (_schema, 4, 'install_tenant_v4_tables')
    ON CONFLICT (schema_name, version) DO NOTHING;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
