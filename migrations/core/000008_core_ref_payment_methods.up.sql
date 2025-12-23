-- +goose Up
-- +goose StatementBegin

-- Payment Methods reference table
CREATE TABLE core_payment_methods (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    code VARCHAR(50) UNIQUE NOT NULL, -- visa, mastercard, paypal, stripe, crypto_btc, bank_transfer
    name VARCHAR(100) NOT NULL, -- Human readable name
    payment_type VARCHAR(30) NOT NULL, -- card, bank, crypto, digital_wallet, cash, buy_now_pay_later
    icon_url VARCHAR(255), -- URL to payment method icon/logo
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE core_payment_methods IS 'Available payment methods for e-commerce and transactions';
COMMENT ON COLUMN core_payment_methods.code IS 'Unique code identifier (lowercase with underscores)';
COMMENT ON COLUMN core_payment_methods.payment_type IS 'Category: card, bank, crypto, digital_wallet, cash, buy_now_pay_later';
COMMENT ON COLUMN core_payment_methods.icon_url IS 'Optional URL to icon or logo for display';

-- Indexes
CREATE INDEX idx_core_payment_methods_code ON core_payment_methods(code);
CREATE INDEX idx_core_payment_methods_type ON core_payment_methods(payment_type);
CREATE INDEX idx_core_payment_methods_active ON core_payment_methods(is_active);

-- Trigger for updated_at
CREATE TRIGGER trg_core_payment_methods_updated_at
    BEFORE UPDATE ON core_payment_methods
    FOR EACH ROW
    EXECUTE FUNCTION tfn_entity_updated_at();

-- Insert common payment methods
INSERT INTO core_payment_methods (code, name, payment_type, description, sort_order) VALUES
-- Card payments
('visa', 'Visa', 'card', 'Visa credit/debit card', 1),
('mastercard', 'Mastercard', 'card', 'Mastercard credit/debit card', 2),
('amex', 'American Express', 'card', 'American Express credit card', 3),
('discover', 'Discover', 'card', 'Discover credit card', 4),
('maestro', 'Maestro', 'card', 'Maestro debit card', 5),
('jcb', 'JCB', 'card', 'JCB credit card', 6),
('unionpay', 'UnionPay', 'card', 'China UnionPay card', 7),
('mir', 'МИР', 'card', 'Russian МИР payment system', 8),

-- Digital wallets
('paypal', 'PayPal', 'digital_wallet', 'PayPal digital wallet', 10),
('apple_pay', 'Apple Pay', 'digital_wallet', 'Apple Pay mobile payment', 11),
('google_pay', 'Google Pay', 'digital_wallet', 'Google Pay mobile payment', 12),
('samsung_pay', 'Samsung Pay', 'digital_wallet', 'Samsung Pay mobile payment', 13),
('yandex_money', 'ЮMoney', 'digital_wallet', 'Yandex Money (Russia)', 14),
('qiwi', 'QIWI Wallet', 'digital_wallet', 'QIWI digital wallet (Russia)', 15),
('webmoney', 'WebMoney', 'digital_wallet', 'WebMoney payment system', 16),
('alipay', 'Alipay', 'digital_wallet', 'Alipay mobile payment (China)', 17),
('wechat_pay', 'WeChat Pay', 'digital_wallet', 'WeChat Pay mobile payment (China)', 18),

-- Bank transfers
('bank_transfer', 'Bank Transfer', 'bank', 'Direct bank transfer', 20),
('sepa', 'SEPA', 'bank', 'SEPA bank transfer (Europe)', 21),
('ach', 'ACH', 'bank', 'ACH bank transfer (USA)', 22),
('swift', 'SWIFT', 'bank', 'SWIFT international transfer', 23),

-- Cryptocurrency
('crypto_btc', 'Bitcoin', 'crypto', 'Bitcoin cryptocurrency', 30),
('crypto_eth', 'Ethereum', 'crypto', 'Ethereum cryptocurrency', 31),
('crypto_usdt', 'USDT', 'crypto', 'Tether stablecoin', 32),
('crypto_usdc', 'USDC', 'crypto', 'USD Coin stablecoin', 33),

-- Buy Now Pay Later
('klarna', 'Klarna', 'buy_now_pay_later', 'Klarna installment payments', 40),
('afterpay', 'Afterpay', 'buy_now_pay_later', 'Afterpay installment payments', 41),
('affirm', 'Affirm', 'buy_now_pay_later', 'Affirm financing', 42),

-- Payment processors/gateways
('stripe', 'Stripe', 'digital_wallet', 'Stripe payment gateway', 50),
('square', 'Square', 'digital_wallet', 'Square payment processing', 51),
('braintree', 'Braintree', 'digital_wallet', 'Braintree payment gateway', 52),

-- Cash/Other
('cash', 'Cash', 'cash', 'Cash payment', 60),
('cash_on_delivery', 'Cash on Delivery', 'cash', 'Pay with cash upon delivery', 61);

-- +goose StatementEnd
