-- Masukkan semua wallets ke participants
INSERT INTO public.participants (type, ref_id)
SELECT 'wallet', id FROM public.wallets;

-- Masukkan semua internal_accounts ke participants
INSERT INTO public.participants (type, ref_id)
SELECT 'internal', id FROM public.internal_accounts;
