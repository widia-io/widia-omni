SET search_path TO widia_omni;

DELETE FROM plans WHERE tier IN ('free', 'pro', 'premium');
