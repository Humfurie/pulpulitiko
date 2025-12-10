-- Migration: 000009_location_seed_data
-- Seed data for Philippine provinces, cities/municipalities, and congressional districts
-- Based on PSGC (Philippine Standard Geographic Code) 2023

-- =====================================================
-- PROVINCES
-- =====================================================

-- Region I - Ilocos Region
INSERT INTO provinces (region_id, code, name, slug) VALUES
((SELECT id FROM regions WHERE code = '010000000'), '012800000', 'Ilocos Norte', 'ilocos-norte'),
((SELECT id FROM regions WHERE code = '010000000'), '012900000', 'Ilocos Sur', 'ilocos-sur'),
((SELECT id FROM regions WHERE code = '010000000'), '013300000', 'La Union', 'la-union'),
((SELECT id FROM regions WHERE code = '010000000'), '015500000', 'Pangasinan', 'pangasinan');

-- Region II - Cagayan Valley
INSERT INTO provinces (region_id, code, name, slug) VALUES
((SELECT id FROM regions WHERE code = '020000000'), '021500000', 'Cagayan', 'cagayan'),
((SELECT id FROM regions WHERE code = '020000000'), '023100000', 'Isabela', 'isabela'),
((SELECT id FROM regions WHERE code = '020000000'), '025000000', 'Nueva Vizcaya', 'nueva-vizcaya'),
((SELECT id FROM regions WHERE code = '020000000'), '025700000', 'Quirino', 'quirino'),
((SELECT id FROM regions WHERE code = '020000000'), '020900000', 'Batanes', 'batanes');

-- Region III - Central Luzon
INSERT INTO provinces (region_id, code, name, slug) VALUES
((SELECT id FROM regions WHERE code = '030000000'), '030800000', 'Bataan', 'bataan'),
((SELECT id FROM regions WHERE code = '030000000'), '031400000', 'Bulacan', 'bulacan'),
((SELECT id FROM regions WHERE code = '030000000'), '034900000', 'Nueva Ecija', 'nueva-ecija'),
((SELECT id FROM regions WHERE code = '030000000'), '035400000', 'Pampanga', 'pampanga'),
((SELECT id FROM regions WHERE code = '030000000'), '036900000', 'Tarlac', 'tarlac'),
((SELECT id FROM regions WHERE code = '030000000'), '037100000', 'Zambales', 'zambales'),
((SELECT id FROM regions WHERE code = '030000000'), '037700000', 'Aurora', 'aurora');

-- Region IV-A - CALABARZON
INSERT INTO provinces (region_id, code, name, slug) VALUES
((SELECT id FROM regions WHERE code = '040000000'), '041000000', 'Batangas', 'batangas'),
((SELECT id FROM regions WHERE code = '040000000'), '042100000', 'Cavite', 'cavite'),
((SELECT id FROM regions WHERE code = '040000000'), '043400000', 'Laguna', 'laguna'),
((SELECT id FROM regions WHERE code = '040000000'), '045600000', 'Quezon', 'quezon'),
((SELECT id FROM regions WHERE code = '040000000'), '045800000', 'Rizal', 'rizal');

-- Region IV-B - MIMAROPA
INSERT INTO provinces (region_id, code, name, slug) VALUES
((SELECT id FROM regions WHERE code = '170000000'), '174000000', 'Marinduque', 'marinduque'),
((SELECT id FROM regions WHERE code = '170000000'), '175100000', 'Occidental Mindoro', 'occidental-mindoro'),
((SELECT id FROM regions WHERE code = '170000000'), '175200000', 'Oriental Mindoro', 'oriental-mindoro'),
((SELECT id FROM regions WHERE code = '170000000'), '175300000', 'Palawan', 'palawan'),
((SELECT id FROM regions WHERE code = '170000000'), '175900000', 'Romblon', 'romblon');

-- Region V - Bicol Region
INSERT INTO provinces (region_id, code, name, slug) VALUES
((SELECT id FROM regions WHERE code = '050000000'), '050500000', 'Albay', 'albay'),
((SELECT id FROM regions WHERE code = '050000000'), '051600000', 'Camarines Norte', 'camarines-norte'),
((SELECT id FROM regions WHERE code = '050000000'), '051700000', 'Camarines Sur', 'camarines-sur'),
((SELECT id FROM regions WHERE code = '050000000'), '052000000', 'Catanduanes', 'catanduanes'),
((SELECT id FROM regions WHERE code = '050000000'), '054100000', 'Masbate', 'masbate'),
((SELECT id FROM regions WHERE code = '050000000'), '056200000', 'Sorsogon', 'sorsogon');

-- Region VI - Western Visayas
INSERT INTO provinces (region_id, code, name, slug) VALUES
((SELECT id FROM regions WHERE code = '060000000'), '060400000', 'Aklan', 'aklan'),
((SELECT id FROM regions WHERE code = '060000000'), '060600000', 'Antique', 'antique'),
((SELECT id FROM regions WHERE code = '060000000'), '061900000', 'Capiz', 'capiz'),
((SELECT id FROM regions WHERE code = '060000000'), '063000000', 'Iloilo', 'iloilo'),
((SELECT id FROM regions WHERE code = '060000000'), '064500000', 'Negros Occidental', 'negros-occidental'),
((SELECT id FROM regions WHERE code = '060000000'), '067900000', 'Guimaras', 'guimaras');

-- Region VII - Central Visayas
INSERT INTO provinces (region_id, code, name, slug) VALUES
((SELECT id FROM regions WHERE code = '070000000'), '071200000', 'Bohol', 'bohol'),
((SELECT id FROM regions WHERE code = '070000000'), '072200000', 'Cebu', 'cebu'),
((SELECT id FROM regions WHERE code = '070000000'), '074600000', 'Negros Oriental', 'negros-oriental'),
((SELECT id FROM regions WHERE code = '070000000'), '076100000', 'Siquijor', 'siquijor');

-- Region VIII - Eastern Visayas
INSERT INTO provinces (region_id, code, name, slug) VALUES
((SELECT id FROM regions WHERE code = '080000000'), '082600000', 'Eastern Samar', 'eastern-samar'),
((SELECT id FROM regions WHERE code = '080000000'), '083700000', 'Leyte', 'leyte'),
((SELECT id FROM regions WHERE code = '080000000'), '084800000', 'Northern Samar', 'northern-samar'),
((SELECT id FROM regions WHERE code = '080000000'), '086000000', 'Samar', 'samar'),
((SELECT id FROM regions WHERE code = '080000000'), '086400000', 'Southern Leyte', 'southern-leyte'),
((SELECT id FROM regions WHERE code = '080000000'), '087800000', 'Biliran', 'biliran');

-- Region IX - Zamboanga Peninsula
INSERT INTO provinces (region_id, code, name, slug) VALUES
((SELECT id FROM regions WHERE code = '090000000'), '097200000', 'Zamboanga del Norte', 'zamboanga-del-norte'),
((SELECT id FROM regions WHERE code = '090000000'), '097300000', 'Zamboanga del Sur', 'zamboanga-del-sur'),
((SELECT id FROM regions WHERE code = '090000000'), '098300000', 'Zamboanga Sibugay', 'zamboanga-sibugay');

-- Region X - Northern Mindanao
INSERT INTO provinces (region_id, code, name, slug) VALUES
((SELECT id FROM regions WHERE code = '100000000'), '101300000', 'Bukidnon', 'bukidnon'),
((SELECT id FROM regions WHERE code = '100000000'), '101800000', 'Camiguin', 'camiguin'),
((SELECT id FROM regions WHERE code = '100000000'), '103500000', 'Lanao del Norte', 'lanao-del-norte'),
((SELECT id FROM regions WHERE code = '100000000'), '104200000', 'Misamis Occidental', 'misamis-occidental'),
((SELECT id FROM regions WHERE code = '100000000'), '104300000', 'Misamis Oriental', 'misamis-oriental');

-- Region XI - Davao Region
INSERT INTO provinces (region_id, code, name, slug) VALUES
((SELECT id FROM regions WHERE code = '110000000'), '112300000', 'Davao de Oro', 'davao-de-oro'),
((SELECT id FROM regions WHERE code = '110000000'), '112400000', 'Davao del Norte', 'davao-del-norte'),
((SELECT id FROM regions WHERE code = '110000000'), '112500000', 'Davao del Sur', 'davao-del-sur'),
((SELECT id FROM regions WHERE code = '110000000'), '118200000', 'Davao Oriental', 'davao-oriental'),
((SELECT id FROM regions WHERE code = '110000000'), '118600000', 'Davao Occidental', 'davao-occidental');

-- Region XII - SOCCSKSARGEN
INSERT INTO provinces (region_id, code, name, slug) VALUES
((SELECT id FROM regions WHERE code = '120000000'), '124700000', 'Cotabato', 'cotabato'),
((SELECT id FROM regions WHERE code = '120000000'), '126300000', 'South Cotabato', 'south-cotabato'),
((SELECT id FROM regions WHERE code = '120000000'), '126500000', 'Sultan Kudarat', 'sultan-kudarat'),
((SELECT id FROM regions WHERE code = '120000000'), '128000000', 'Sarangani', 'sarangani');

-- NCR - National Capital Region (Special case: NCR has no provinces, only cities/municipalities)
-- We'll create a placeholder province for NCR districts
INSERT INTO provinces (region_id, code, name, slug) VALUES
((SELECT id FROM regions WHERE code = '130000000'), '133900000', 'Metro Manila', 'metro-manila');

-- CAR - Cordillera Administrative Region
INSERT INTO provinces (region_id, code, name, slug) VALUES
((SELECT id FROM regions WHERE code = '140000000'), '140100000', 'Abra', 'abra'),
((SELECT id FROM regions WHERE code = '140000000'), '141100000', 'Benguet', 'benguet'),
((SELECT id FROM regions WHERE code = '140000000'), '142700000', 'Ifugao', 'ifugao'),
((SELECT id FROM regions WHERE code = '140000000'), '143200000', 'Kalinga', 'kalinga'),
((SELECT id FROM regions WHERE code = '140000000'), '144400000', 'Mountain Province', 'mountain-province'),
((SELECT id FROM regions WHERE code = '140000000'), '148100000', 'Apayao', 'apayao');

-- Region XIII - Caraga
INSERT INTO provinces (region_id, code, name, slug) VALUES
((SELECT id FROM regions WHERE code = '160000000'), '160200000', 'Agusan del Norte', 'agusan-del-norte'),
((SELECT id FROM regions WHERE code = '160000000'), '160300000', 'Agusan del Sur', 'agusan-del-sur'),
((SELECT id FROM regions WHERE code = '160000000'), '166700000', 'Surigao del Norte', 'surigao-del-norte'),
((SELECT id FROM regions WHERE code = '160000000'), '166800000', 'Surigao del Sur', 'surigao-del-sur'),
((SELECT id FROM regions WHERE code = '160000000'), '168500000', 'Dinagat Islands', 'dinagat-islands');

-- BARMM - Bangsamoro Autonomous Region in Muslim Mindanao
INSERT INTO provinces (region_id, code, name, slug) VALUES
((SELECT id FROM regions WHERE code = '190000000'), '150700000', 'Basilan', 'basilan'),
((SELECT id FROM regions WHERE code = '190000000'), '153600000', 'Lanao del Sur', 'lanao-del-sur'),
((SELECT id FROM regions WHERE code = '190000000'), '153800000', 'Maguindanao del Norte', 'maguindanao-del-norte'),
((SELECT id FROM regions WHERE code = '190000000'), '153900000', 'Maguindanao del Sur', 'maguindanao-del-sur'),
((SELECT id FROM regions WHERE code = '190000000'), '156600000', 'Sulu', 'sulu'),
((SELECT id FROM regions WHERE code = '190000000'), '157000000', 'Tawi-Tawi', 'tawi-tawi');

-- =====================================================
-- CITIES AND MUNICIPALITIES (Major ones per province)
-- =====================================================

-- NCR Cities (Metro Manila)
INSERT INTO cities_municipalities (province_id, code, name, slug, is_city, is_huc) VALUES
((SELECT id FROM provinces WHERE code = '133900000'), '133900100', 'Manila', 'manila', TRUE, TRUE),
((SELECT id FROM provinces WHERE code = '133900000'), '137400000', 'Quezon City', 'quezon-city', TRUE, TRUE),
((SELECT id FROM provinces WHERE code = '133900000'), '137500000', 'Caloocan', 'caloocan', TRUE, TRUE),
((SELECT id FROM provinces WHERE code = '133900000'), '137600000', 'Las Piñas', 'las-pinas', TRUE, TRUE),
((SELECT id FROM provinces WHERE code = '133900000'), '137700000', 'Makati', 'makati', TRUE, TRUE),
((SELECT id FROM provinces WHERE code = '133900000'), '137800000', 'Malabon', 'malabon', TRUE, TRUE),
((SELECT id FROM provinces WHERE code = '133900000'), '137900000', 'Mandaluyong', 'mandaluyong', TRUE, TRUE),
((SELECT id FROM provinces WHERE code = '133900000'), '138000000', 'Marikina', 'marikina', TRUE, TRUE),
((SELECT id FROM provinces WHERE code = '133900000'), '138100000', 'Muntinlupa', 'muntinlupa', TRUE, TRUE),
((SELECT id FROM provinces WHERE code = '133900000'), '138200000', 'Navotas', 'navotas', TRUE, TRUE),
((SELECT id FROM provinces WHERE code = '133900000'), '138300000', 'Parañaque', 'paranaque', TRUE, TRUE),
((SELECT id FROM provinces WHERE code = '133900000'), '138400000', 'Pasay', 'pasay', TRUE, TRUE),
((SELECT id FROM provinces WHERE code = '133900000'), '138500000', 'Pasig', 'pasig', TRUE, TRUE),
((SELECT id FROM provinces WHERE code = '133900000'), '138600000', 'San Juan', 'san-juan', TRUE, TRUE),
((SELECT id FROM provinces WHERE code = '133900000'), '138700000', 'Taguig', 'taguig', TRUE, TRUE),
((SELECT id FROM provinces WHERE code = '133900000'), '138800000', 'Valenzuela', 'valenzuela', TRUE, TRUE),
((SELECT id FROM provinces WHERE code = '133900000'), '137402000', 'Pateros', 'pateros', FALSE, FALSE);

-- Cebu Province - Major Cities
INSERT INTO cities_municipalities (province_id, code, name, slug, is_city, is_huc) VALUES
((SELECT id FROM provinces WHERE code = '072200000'), '072217000', 'Cebu City', 'cebu-city', TRUE, TRUE),
((SELECT id FROM provinces WHERE code = '072200000'), '072230000', 'Mandaue', 'mandaue', TRUE, TRUE),
((SELECT id FROM provinces WHERE code = '072200000'), '072233000', 'Lapu-Lapu', 'lapu-lapu', TRUE, TRUE),
((SELECT id FROM provinces WHERE code = '072200000'), '072263000', 'Talisay', 'talisay-cebu', TRUE, FALSE),
((SELECT id FROM provinces WHERE code = '072200000'), '072243000', 'Naga', 'naga-cebu', TRUE, FALSE),
((SELECT id FROM provinces WHERE code = '072200000'), '072216000', 'Carcar', 'carcar', TRUE, FALSE),
((SELECT id FROM provinces WHERE code = '072200000'), '072219000', 'Danao', 'danao', TRUE, FALSE),
((SELECT id FROM provinces WHERE code = '072200000'), '072270000', 'Toledo', 'toledo', TRUE, FALSE),
((SELECT id FROM provinces WHERE code = '072200000'), '072211000', 'Bogo', 'bogo', TRUE, FALSE);

-- Davao del Sur - Major Cities
INSERT INTO cities_municipalities (province_id, code, name, slug, is_city, is_huc) VALUES
((SELECT id FROM provinces WHERE code = '112500000'), '112402000', 'Davao City', 'davao-city', TRUE, TRUE),
((SELECT id FROM provinces WHERE code = '112500000'), '112503000', 'Digos', 'digos', TRUE, FALSE);

-- Iloilo Province - Major Cities
INSERT INTO cities_municipalities (province_id, code, name, slug, is_city, is_huc) VALUES
((SELECT id FROM provinces WHERE code = '063000000'), '063000000', 'Iloilo City', 'iloilo-city', TRUE, TRUE),
((SELECT id FROM provinces WHERE code = '063000000'), '063045000', 'Passi', 'passi', TRUE, FALSE);

-- Bulacan Province - Major Cities
INSERT INTO cities_municipalities (province_id, code, name, slug, is_city, is_huc) VALUES
((SELECT id FROM provinces WHERE code = '031400000'), '031410000', 'Malolos', 'malolos', TRUE, FALSE),
((SELECT id FROM provinces WHERE code = '031400000'), '031414000', 'Meycauayan', 'meycauayan', TRUE, FALSE),
((SELECT id FROM provinces WHERE code = '031400000'), '031426000', 'San Jose del Monte', 'san-jose-del-monte', TRUE, FALSE);

-- Cavite Province - Major Cities
INSERT INTO cities_municipalities (province_id, code, name, slug, is_city, is_huc) VALUES
((SELECT id FROM provinces WHERE code = '042100000'), '042108000', 'Cavite City', 'cavite-city', TRUE, FALSE),
((SELECT id FROM provinces WHERE code = '042100000'), '042106000', 'Bacoor', 'bacoor', TRUE, FALSE),
((SELECT id FROM provinces WHERE code = '042100000'), '042109000', 'Dasmariñas', 'dasmarinas', TRUE, FALSE),
((SELECT id FROM provinces WHERE code = '042100000'), '042110000', 'Imus', 'imus', TRUE, FALSE),
((SELECT id FROM provinces WHERE code = '042100000'), '042123000', 'Tagaytay', 'tagaytay', TRUE, FALSE),
((SELECT id FROM provinces WHERE code = '042100000'), '042125000', 'Trece Martires', 'trece-martires', TRUE, FALSE),
((SELECT id FROM provinces WHERE code = '042100000'), '042111000', 'General Trias', 'general-trias', TRUE, FALSE);

-- Laguna Province - Major Cities
INSERT INTO cities_municipalities (province_id, code, name, slug, is_city, is_huc) VALUES
((SELECT id FROM provinces WHERE code = '043400000'), '043404000', 'Biñan', 'binan', TRUE, FALSE),
((SELECT id FROM provinces WHERE code = '043400000'), '043405000', 'Cabuyao', 'cabuyao', TRUE, FALSE),
((SELECT id FROM provinces WHERE code = '043400000'), '043406000', 'Calamba', 'calamba', TRUE, FALSE),
((SELECT id FROM provinces WHERE code = '043400000'), '043424000', 'San Pablo', 'san-pablo', TRUE, FALSE),
((SELECT id FROM provinces WHERE code = '043400000'), '043425000', 'San Pedro', 'san-pedro', TRUE, FALSE),
((SELECT id FROM provinces WHERE code = '043400000'), '043426000', 'Santa Rosa', 'santa-rosa-laguna', TRUE, FALSE);

-- Batangas Province - Major Cities
INSERT INTO cities_municipalities (province_id, code, name, slug, is_city, is_huc) VALUES
((SELECT id FROM provinces WHERE code = '041000000'), '041005000', 'Batangas City', 'batangas-city', TRUE, FALSE),
((SELECT id FROM provinces WHERE code = '041000000'), '041017000', 'Lipa', 'lipa', TRUE, FALSE),
((SELECT id FROM provinces WHERE code = '041000000'), '041030000', 'Tanauan', 'tanauan', TRUE, FALSE),
((SELECT id FROM provinces WHERE code = '041000000'), '041027000', 'Santo Tomas', 'santo-tomas-batangas', TRUE, FALSE);

-- Pampanga Province - Major Cities
INSERT INTO cities_municipalities (province_id, code, name, slug, is_city, is_huc) VALUES
((SELECT id FROM provinces WHERE code = '035400000'), '035401000', 'Angeles', 'angeles', TRUE, TRUE),
((SELECT id FROM provinces WHERE code = '035400000'), '035419000', 'San Fernando', 'san-fernando-pampanga', TRUE, FALSE),
((SELECT id FROM provinces WHERE code = '035400000'), '035411000', 'Mabalacat', 'mabalacat', TRUE, FALSE);

-- Negros Occidental - Major Cities
INSERT INTO cities_municipalities (province_id, code, name, slug, is_city, is_huc) VALUES
((SELECT id FROM provinces WHERE code = '064500000'), '064502000', 'Bacolod', 'bacolod', TRUE, TRUE),
((SELECT id FROM provinces WHERE code = '064500000'), '064503000', 'Bago', 'bago', TRUE, FALSE),
((SELECT id FROM provinces WHERE code = '064500000'), '064506000', 'Cadiz', 'cadiz', TRUE, FALSE),
((SELECT id FROM provinces WHERE code = '064500000'), '064522000', 'San Carlos', 'san-carlos-negros-occidental', TRUE, FALSE),
((SELECT id FROM provinces WHERE code = '064500000'), '064524000', 'Silay', 'silay', TRUE, FALSE),
((SELECT id FROM provinces WHERE code = '064500000'), '064530000', 'Victorias', 'victorias', TRUE, FALSE);

-- Pangasinan - Major Cities
INSERT INTO cities_municipalities (province_id, code, name, slug, is_city, is_huc) VALUES
((SELECT id FROM provinces WHERE code = '015500000'), '015503000', 'Alaminos', 'alaminos', TRUE, FALSE),
((SELECT id FROM provinces WHERE code = '015500000'), '015518000', 'Dagupan', 'dagupan', TRUE, FALSE),
((SELECT id FROM provinces WHERE code = '015500000'), '015544000', 'San Carlos', 'san-carlos-pangasinan', TRUE, FALSE),
((SELECT id FROM provinces WHERE code = '015500000'), '015586000', 'Urdaneta', 'urdaneta', TRUE, FALSE);

-- Zambales - Major Cities/Municipalities
INSERT INTO cities_municipalities (province_id, code, name, slug, is_city, is_huc) VALUES
((SELECT id FROM provinces WHERE code = '037100000'), '037108000', 'Olongapo', 'olongapo', TRUE, TRUE);

-- Benguet - Major Cities
INSERT INTO cities_municipalities (province_id, code, name, slug, is_city, is_huc) VALUES
((SELECT id FROM provinces WHERE code = '141100000'), '141102000', 'Baguio', 'baguio', TRUE, TRUE);

-- Leyte - Major Cities
INSERT INTO cities_municipalities (province_id, code, name, slug, is_city, is_huc) VALUES
((SELECT id FROM provinces WHERE code = '083700000'), '083747000', 'Tacloban', 'tacloban', TRUE, TRUE),
((SELECT id FROM provinces WHERE code = '083700000'), '083745000', 'Ormoc', 'ormoc', TRUE, FALSE);

-- South Cotabato - Major Cities
INSERT INTO cities_municipalities (province_id, code, name, slug, is_city, is_huc) VALUES
((SELECT id FROM provinces WHERE code = '126300000'), '126303000', 'General Santos', 'general-santos', TRUE, TRUE),
((SELECT id FROM provinces WHERE code = '126300000'), '126306000', 'Koronadal', 'koronadal', TRUE, FALSE);

-- Cagayan de Oro (Misamis Oriental)
INSERT INTO cities_municipalities (province_id, code, name, slug, is_city, is_huc) VALUES
((SELECT id FROM provinces WHERE code = '104300000'), '104305000', 'Cagayan de Oro', 'cagayan-de-oro', TRUE, TRUE);

-- Zamboanga del Sur
INSERT INTO cities_municipalities (province_id, code, name, slug, is_city, is_huc) VALUES
((SELECT id FROM provinces WHERE code = '097300000'), '097332000', 'Zamboanga City', 'zamboanga-city', TRUE, TRUE),
((SELECT id FROM provinces WHERE code = '097300000'), '097322000', 'Pagadian', 'pagadian', TRUE, FALSE);

-- Lanao del Sur
INSERT INTO cities_municipalities (province_id, code, name, slug, is_city, is_huc) VALUES
((SELECT id FROM provinces WHERE code = '153600000'), '153604000', 'Marawi', 'marawi', TRUE, FALSE);

-- Isabela
INSERT INTO cities_municipalities (province_id, code, name, slug, is_city, is_huc) VALUES
((SELECT id FROM provinces WHERE code = '023100000'), '023114000', 'Cauayan', 'cauayan', TRUE, FALSE),
((SELECT id FROM provinces WHERE code = '023100000'), '023128000', 'Ilagan', 'ilagan', TRUE, FALSE),
((SELECT id FROM provinces WHERE code = '023100000'), '023137000', 'Santiago', 'santiago', TRUE, FALSE);

-- Nueva Ecija
INSERT INTO cities_municipalities (province_id, code, name, slug, is_city, is_huc) VALUES
((SELECT id FROM provinces WHERE code = '034900000'), '034903000', 'Cabanatuan', 'cabanatuan', TRUE, FALSE),
((SELECT id FROM provinces WHERE code = '034900000'), '034919000', 'Palayan', 'palayan', TRUE, FALSE),
((SELECT id FROM provinces WHERE code = '034900000'), '034924000', 'San Jose', 'san-jose-nueva-ecija', TRUE, FALSE),
((SELECT id FROM provinces WHERE code = '034900000'), '034912000', 'Gapan', 'gapan', TRUE, FALSE),
((SELECT id FROM provinces WHERE code = '034900000'), '034927000', 'Science City of Muñoz', 'munoz', TRUE, FALSE);

-- Tarlac
INSERT INTO cities_municipalities (province_id, code, name, slug, is_city, is_huc) VALUES
((SELECT id FROM provinces WHERE code = '036900000'), '036916000', 'Tarlac City', 'tarlac-city', TRUE, FALSE);

-- Rizal
INSERT INTO cities_municipalities (province_id, code, name, slug, is_city, is_huc) VALUES
((SELECT id FROM provinces WHERE code = '045800000'), '045802000', 'Antipolo', 'antipolo', TRUE, TRUE);

-- Albay
INSERT INTO cities_municipalities (province_id, code, name, slug, is_city, is_huc) VALUES
((SELECT id FROM provinces WHERE code = '050500000'), '050506000', 'Legazpi', 'legazpi', TRUE, FALSE),
((SELECT id FROM provinces WHERE code = '050500000'), '050515000', 'Ligao', 'ligao', TRUE, FALSE),
((SELECT id FROM provinces WHERE code = '050500000'), '050524000', 'Tabaco', 'tabaco', TRUE, FALSE);

-- Camarines Sur
INSERT INTO cities_municipalities (province_id, code, name, slug, is_city, is_huc) VALUES
((SELECT id FROM provinces WHERE code = '051700000'), '051711000', 'Iriga', 'iriga', TRUE, FALSE),
((SELECT id FROM provinces WHERE code = '051700000'), '051720000', 'Naga', 'naga-camarines-sur', TRUE, FALSE);

-- Sorsogon
INSERT INTO cities_municipalities (province_id, code, name, slug, is_city, is_huc) VALUES
((SELECT id FROM provinces WHERE code = '056200000'), '056214000', 'Sorsogon City', 'sorsogon-city', TRUE, FALSE);

-- Bohol
INSERT INTO cities_municipalities (province_id, code, name, slug, is_city, is_huc) VALUES
((SELECT id FROM provinces WHERE code = '071200000'), '071246000', 'Tagbilaran', 'tagbilaran', TRUE, FALSE);

-- Negros Oriental
INSERT INTO cities_municipalities (province_id, code, name, slug, is_city, is_huc) VALUES
((SELECT id FROM provinces WHERE code = '074600000'), '074605000', 'Bais', 'bais', TRUE, FALSE),
((SELECT id FROM provinces WHERE code = '074600000'), '074608000', 'Bayawan', 'bayawan', TRUE, FALSE),
((SELECT id FROM provinces WHERE code = '074600000'), '074610000', 'Canlaon', 'canlaon', TRUE, FALSE),
((SELECT id FROM provinces WHERE code = '074600000'), '074612000', 'Dumaguete', 'dumaguete', TRUE, FALSE),
((SELECT id FROM provinces WHERE code = '074600000'), '074616000', 'Guihulngan', 'guihulngan', TRUE, FALSE),
((SELECT id FROM provinces WHERE code = '074600000'), '074625000', 'Tanjay', 'tanjay', TRUE, FALSE);

-- =====================================================
-- CONGRESSIONAL DISTRICTS (Sample - Major Districts)
-- =====================================================

-- NCR Congressional Districts
INSERT INTO congressional_districts (city_municipality_id, district_number, name, slug) VALUES
((SELECT id FROM cities_municipalities WHERE code = '133900100'), 1, '1st District of Manila', 'manila-1st-district'),
((SELECT id FROM cities_municipalities WHERE code = '133900100'), 2, '2nd District of Manila', 'manila-2nd-district'),
((SELECT id FROM cities_municipalities WHERE code = '133900100'), 3, '3rd District of Manila', 'manila-3rd-district'),
((SELECT id FROM cities_municipalities WHERE code = '133900100'), 4, '4th District of Manila', 'manila-4th-district'),
((SELECT id FROM cities_municipalities WHERE code = '133900100'), 5, '5th District of Manila', 'manila-5th-district'),
((SELECT id FROM cities_municipalities WHERE code = '133900100'), 6, '6th District of Manila', 'manila-6th-district'),
((SELECT id FROM cities_municipalities WHERE code = '137400000'), 1, '1st District of Quezon City', 'quezon-city-1st-district'),
((SELECT id FROM cities_municipalities WHERE code = '137400000'), 2, '2nd District of Quezon City', 'quezon-city-2nd-district'),
((SELECT id FROM cities_municipalities WHERE code = '137400000'), 3, '3rd District of Quezon City', 'quezon-city-3rd-district'),
((SELECT id FROM cities_municipalities WHERE code = '137400000'), 4, '4th District of Quezon City', 'quezon-city-4th-district'),
((SELECT id FROM cities_municipalities WHERE code = '137400000'), 5, '5th District of Quezon City', 'quezon-city-5th-district'),
((SELECT id FROM cities_municipalities WHERE code = '137400000'), 6, '6th District of Quezon City', 'quezon-city-6th-district');

-- Cebu Congressional Districts
INSERT INTO congressional_districts (province_id, district_number, name, slug) VALUES
((SELECT id FROM provinces WHERE code = '072200000'), 1, '1st District of Cebu', 'cebu-1st-district'),
((SELECT id FROM provinces WHERE code = '072200000'), 2, '2nd District of Cebu', 'cebu-2nd-district'),
((SELECT id FROM provinces WHERE code = '072200000'), 3, '3rd District of Cebu', 'cebu-3rd-district'),
((SELECT id FROM provinces WHERE code = '072200000'), 4, '4th District of Cebu', 'cebu-4th-district'),
((SELECT id FROM provinces WHERE code = '072200000'), 5, '5th District of Cebu', 'cebu-5th-district'),
((SELECT id FROM provinces WHERE code = '072200000'), 6, '6th District of Cebu', 'cebu-6th-district'),
((SELECT id FROM provinces WHERE code = '072200000'), 7, '7th District of Cebu', 'cebu-7th-district');

-- HUC Lone Districts
INSERT INTO congressional_districts (city_municipality_id, district_number, name, slug) VALUES
((SELECT id FROM cities_municipalities WHERE code = '072217000'), 1, 'North District of Cebu City', 'cebu-city-north-district'),
((SELECT id FROM cities_municipalities WHERE code = '072217000'), 2, 'South District of Cebu City', 'cebu-city-south-district'),
((SELECT id FROM cities_municipalities WHERE code = '112402000'), 1, '1st District of Davao City', 'davao-city-1st-district'),
((SELECT id FROM cities_municipalities WHERE code = '112402000'), 2, '2nd District of Davao City', 'davao-city-2nd-district'),
((SELECT id FROM cities_municipalities WHERE code = '112402000'), 3, '3rd District of Davao City', 'davao-city-3rd-district'),
((SELECT id FROM cities_municipalities WHERE code = '064502000'), 1, 'Lone District of Bacolod', 'bacolod-lone-district'),
((SELECT id FROM cities_municipalities WHERE code = '063000000'), 1, 'Lone District of Iloilo City', 'iloilo-city-lone-district'),
((SELECT id FROM cities_municipalities WHERE code = '097332000'), 1, '1st District of Zamboanga City', 'zamboanga-city-1st-district'),
((SELECT id FROM cities_municipalities WHERE code = '097332000'), 2, '2nd District of Zamboanga City', 'zamboanga-city-2nd-district'),
((SELECT id FROM cities_municipalities WHERE code = '126303000'), 1, 'Lone District of General Santos', 'gensan-lone-district'),
((SELECT id FROM cities_municipalities WHERE code = '104305000'), 1, '1st District of Cagayan de Oro', 'cdo-1st-district'),
((SELECT id FROM cities_municipalities WHERE code = '104305000'), 2, '2nd District of Cagayan de Oro', 'cdo-2nd-district'),
((SELECT id FROM cities_municipalities WHERE code = '141102000'), 1, 'Lone District of Baguio', 'baguio-lone-district'),
((SELECT id FROM cities_municipalities WHERE code = '035401000'), 1, 'Lone District of Angeles', 'angeles-lone-district'),
((SELECT id FROM cities_municipalities WHERE code = '037108000'), 1, 'Lone District of Olongapo', 'olongapo-lone-district'),
((SELECT id FROM cities_municipalities WHERE code = '045802000'), 1, '1st District of Antipolo', 'antipolo-1st-district'),
((SELECT id FROM cities_municipalities WHERE code = '045802000'), 2, '2nd District of Antipolo', 'antipolo-2nd-district');

-- Provincial Districts
INSERT INTO congressional_districts (province_id, district_number, name, slug) VALUES
((SELECT id FROM provinces WHERE code = '031400000'), 1, '1st District of Bulacan', 'bulacan-1st-district'),
((SELECT id FROM provinces WHERE code = '031400000'), 2, '2nd District of Bulacan', 'bulacan-2nd-district'),
((SELECT id FROM provinces WHERE code = '031400000'), 3, '3rd District of Bulacan', 'bulacan-3rd-district'),
((SELECT id FROM provinces WHERE code = '031400000'), 4, '4th District of Bulacan', 'bulacan-4th-district'),
((SELECT id FROM provinces WHERE code = '031400000'), 5, '5th District of Bulacan', 'bulacan-5th-district'),
((SELECT id FROM provinces WHERE code = '031400000'), 6, '6th District of Bulacan', 'bulacan-6th-district'),
((SELECT id FROM provinces WHERE code = '042100000'), 1, '1st District of Cavite', 'cavite-1st-district'),
((SELECT id FROM provinces WHERE code = '042100000'), 2, '2nd District of Cavite', 'cavite-2nd-district'),
((SELECT id FROM provinces WHERE code = '042100000'), 3, '3rd District of Cavite', 'cavite-3rd-district'),
((SELECT id FROM provinces WHERE code = '042100000'), 4, '4th District of Cavite', 'cavite-4th-district'),
((SELECT id FROM provinces WHERE code = '042100000'), 5, '5th District of Cavite', 'cavite-5th-district'),
((SELECT id FROM provinces WHERE code = '042100000'), 6, '6th District of Cavite', 'cavite-6th-district'),
((SELECT id FROM provinces WHERE code = '042100000'), 7, '7th District of Cavite', 'cavite-7th-district'),
((SELECT id FROM provinces WHERE code = '043400000'), 1, '1st District of Laguna', 'laguna-1st-district'),
((SELECT id FROM provinces WHERE code = '043400000'), 2, '2nd District of Laguna', 'laguna-2nd-district'),
((SELECT id FROM provinces WHERE code = '043400000'), 3, '3rd District of Laguna', 'laguna-3rd-district'),
((SELECT id FROM provinces WHERE code = '043400000'), 4, '4th District of Laguna', 'laguna-4th-district'),
((SELECT id FROM provinces WHERE code = '041000000'), 1, '1st District of Batangas', 'batangas-1st-district'),
((SELECT id FROM provinces WHERE code = '041000000'), 2, '2nd District of Batangas', 'batangas-2nd-district'),
((SELECT id FROM provinces WHERE code = '041000000'), 3, '3rd District of Batangas', 'batangas-3rd-district'),
((SELECT id FROM provinces WHERE code = '041000000'), 4, '4th District of Batangas', 'batangas-4th-district'),
((SELECT id FROM provinces WHERE code = '041000000'), 5, '5th District of Batangas', 'batangas-5th-district'),
((SELECT id FROM provinces WHERE code = '041000000'), 6, '6th District of Batangas', 'batangas-6th-district'),
((SELECT id FROM provinces WHERE code = '035400000'), 1, '1st District of Pampanga', 'pampanga-1st-district'),
((SELECT id FROM provinces WHERE code = '035400000'), 2, '2nd District of Pampanga', 'pampanga-2nd-district'),
((SELECT id FROM provinces WHERE code = '035400000'), 3, '3rd District of Pampanga', 'pampanga-3rd-district'),
((SELECT id FROM provinces WHERE code = '035400000'), 4, '4th District of Pampanga', 'pampanga-4th-district'),
((SELECT id FROM provinces WHERE code = '015500000'), 1, '1st District of Pangasinan', 'pangasinan-1st-district'),
((SELECT id FROM provinces WHERE code = '015500000'), 2, '2nd District of Pangasinan', 'pangasinan-2nd-district'),
((SELECT id FROM provinces WHERE code = '015500000'), 3, '3rd District of Pangasinan', 'pangasinan-3rd-district'),
((SELECT id FROM provinces WHERE code = '015500000'), 4, '4th District of Pangasinan', 'pangasinan-4th-district'),
((SELECT id FROM provinces WHERE code = '015500000'), 5, '5th District of Pangasinan', 'pangasinan-5th-district'),
((SELECT id FROM provinces WHERE code = '015500000'), 6, '6th District of Pangasinan', 'pangasinan-6th-district'),
((SELECT id FROM provinces WHERE code = '064500000'), 1, '1st District of Negros Occidental', 'negros-occidental-1st-district'),
((SELECT id FROM provinces WHERE code = '064500000'), 2, '2nd District of Negros Occidental', 'negros-occidental-2nd-district'),
((SELECT id FROM provinces WHERE code = '064500000'), 3, '3rd District of Negros Occidental', 'negros-occidental-3rd-district'),
((SELECT id FROM provinces WHERE code = '064500000'), 4, '4th District of Negros Occidental', 'negros-occidental-4th-district'),
((SELECT id FROM provinces WHERE code = '064500000'), 5, '5th District of Negros Occidental', 'negros-occidental-5th-district'),
((SELECT id FROM provinces WHERE code = '034900000'), 1, '1st District of Nueva Ecija', 'nueva-ecija-1st-district'),
((SELECT id FROM provinces WHERE code = '034900000'), 2, '2nd District of Nueva Ecija', 'nueva-ecija-2nd-district'),
((SELECT id FROM provinces WHERE code = '034900000'), 3, '3rd District of Nueva Ecija', 'nueva-ecija-3rd-district'),
((SELECT id FROM provinces WHERE code = '034900000'), 4, '4th District of Nueva Ecija', 'nueva-ecija-4th-district');
