-- Migration 013 — Finance & Accounting: Chart of Account.
-- Tiga tabel: klasifikasi (7, tetap), tipe akun (tetap), dan akun (m_coa).
-- Klasifikasi & saldo normal akun mengikuti tipe; kode akun = kode tipe + urutan.
-- Mengikuti pola master lain (uuid, kolom audit, trigger modified_at dari migration 002).

CREATE TABLE IF NOT EXISTS m_coa_classification (
    classification_code  int  PRIMARY KEY,
    classification_name  text NOT NULL,
    report_name          text NOT NULL   -- 'Balance Sheet' | 'Profit Loss'
);

CREATE TABLE IF NOT EXISTS m_coa_type (
    account_type_code    text PRIMARY KEY,   -- 2 digit, digit-1 = classification_code
    classification_code  int  NOT NULL REFERENCES m_coa_classification(classification_code) ON DELETE RESTRICT,
    account_type_name    text NOT NULL,
    is_credit            boolean NOT NULL DEFAULT false   -- saldo normal default tipe (kontra membaliknya)
);

CREATE TABLE IF NOT EXISTS m_coa (
    id                 uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_code       text NOT NULL UNIQUE,
    account_type_code  text NOT NULL REFERENCES m_coa_type(account_type_code) ON DELETE RESTRICT,
    account_name       text NOT NULL,
    is_contra          boolean NOT NULL DEFAULT false,
    is_credit_account  boolean NOT NULL DEFAULT false,   -- = tipe.is_credit XOR is_contra
    opening_balance    numeric(18,2) NOT NULL DEFAULT 0,
    opening_date       date,
    notes              text,
    is_active          boolean NOT NULL DEFAULT true,
    created_at         timestamptz NOT NULL DEFAULT now(),
    created_by         uuid REFERENCES m_internal_user(id) ON DELETE SET NULL,
    modified_at        timestamptz NOT NULL DEFAULT now(),
    modified_by        uuid REFERENCES m_internal_user(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_coa_type ON m_coa (account_type_code);
CREATE INDEX IF NOT EXISTS idx_coa_type_classification ON m_coa_type (classification_code);

-- Trigger modified_at (fungsi update_modified_at_column dibuat di migration 002).
DROP TRIGGER IF EXISTS update_m_coa_modified_at ON m_coa;
CREATE TRIGGER update_m_coa_modified_at BEFORE UPDATE ON m_coa FOR EACH ROW EXECUTE FUNCTION update_modified_at_column();
