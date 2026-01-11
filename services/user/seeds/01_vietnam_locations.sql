-- Seed file for Vietnam administrative divisions
-- This is a sample with major cities/provinces. 
-- For complete data, see: https://github.com/kenzouno1/DiaGioiHanhChinhVN

-- ============================================
-- CITIES / PROVINCES (Tỉnh/Thành phố)
-- ============================================
-- Major cities first
INSERT INTO cities
  (id, name, code)
VALUES
  (1, 'Hà Nội', '01'),
  (79, 'Thành phố Hồ Chí Minh', '79'),
  (48, 'Đà Nẵng', '48'),
  (31, 'Hải Phòng', '31'),
  (92, 'Cần Thơ', '92');

-- Northern provinces
INSERT INTO cities
  (id, name, code)
VALUES
  (2, 'Hà Giang', '02'),
  (4, 'Cao Bằng', '04'),
  (6, 'Bắc Kạn', '06'),
  (8, 'Tuyên Quang', '08'),
  (10, 'Lào Cai', '10'),
  (11, 'Điện Biên', '11'),
  (12, 'Lai Châu', '12'),
  (14, 'Sơn La', '14'),
  (15, 'Yên Bái', '15'),
  (17, 'Hòa Bình', '17'),
  (19, 'Thái Nguyên', '19'),
  (20, 'Lạng Sơn', '20'),
  (22, 'Quảng Ninh', '22'),
  (24, 'Bắc Giang', '24'),
  (25, 'Phú Thọ', '25'),
  (26, 'Vĩnh Phúc', '26'),
  (27, 'Bắc Ninh', '27'),
  (30, 'Hải Dương', '30'),
  (33, 'Hưng Yên', '33'),
  (34, 'Thái Bình', '34'),
  (35, 'Hà Nam', '35'),
  (36, 'Nam Định', '36'),
  (37, 'Ninh Bình', '37');

-- Central provinces
INSERT INTO cities
  (id, name, code)
VALUES
  (38, 'Thanh Hóa', '38'),
  (40, 'Nghệ An', '40'),
  (42, 'Hà Tĩnh', '42'),
  (44, 'Quảng Bình', '44'),
  (45, 'Quảng Trị', '45'),
  (46, 'Thừa Thiên Huế', '46'),
  (49, 'Quảng Nam', '49'),
  (51, 'Quảng Ngãi', '51'),
  (52, 'Bình Định', '52'),
  (54, 'Phú Yên', '54'),
  (56, 'Khánh Hòa', '56'),
  (58, 'Ninh Thuận', '58'),
  (60, 'Bình Thuận', '60');

-- Highland provinces
INSERT INTO cities
  (id, name, code)
VALUES
  (62, 'Kon Tum', '62'),
  (64, 'Gia Lai', '64'),
  (66, 'Đắk Lắk', '66'),
  (67, 'Đắk Nông', '67'),
  (68, 'Lâm Đồng', '68');

-- Southern provinces
INSERT INTO cities
  (id, name, code)
VALUES
  (70, 'Bình Phước', '70'),
  (72, 'Tây Ninh', '72'),
  (74, 'Bình Dương', '74'),
  (75, 'Đồng Nai', '75'),
  (77, 'Bà Rịa - Vũng Tàu', '77'),
  (80, 'Long An', '80'),
  (82, 'Tiền Giang', '82'),
  (83, 'Bến Tre', '83'),
  (84, 'Trà Vinh', '84'),
  (86, 'Vĩnh Long', '86'),
  (87, 'Đồng Tháp', '87'),
  (89, 'An Giang', '89'),
  (91, 'Kiên Giang', '91'),
  (93, 'Hậu Giang', '93'),
  (94, 'Sóc Trăng', '94'),
  (95, 'Bạc Liêu', '95'),
  (96, 'Cà Mau', '96');

-- ============================================
-- SAMPLE DISTRICTS (Quận/Huyện)
-- ============================================
-- Hanoi Districts
INSERT INTO districts
  (id, city_id, name, code)
VALUES
  (1, 1, 'Ba Đình', '001'),
  (2, 1, 'Hoàn Kiếm', '002'),
  (3, 1, 'Tây Hồ', '003'),
  (4, 1, 'Long Biên', '004'),
  (5, 1, 'Cầu Giấy', '005'),
  (6, 1, 'Đống Đa', '006'),
  (7, 1, 'Hai Bà Trưng', '007'),
  (8, 1, 'Hoàng Mai', '008'),
  (9, 1, 'Thanh Xuân', '009'),
  (10, 1, 'Hà Đông', '010');

-- Ho Chi Minh City Districts
INSERT INTO districts
  (id, city_id, name, code)
VALUES
  (760, 79, 'Quận 1', '760'),
  (761, 79, 'Quận 2', '761'),
  (762, 79, 'Quận 3', '762'),
  (763, 79, 'Quận 4', '763'),
  (764, 79, 'Quận 5', '764'),
  (765, 79, 'Quận 6', '765'),
  (766, 79, 'Quận 7', '766'),
  (767, 79, 'Quận 8', '767'),
  (768, 79, 'Quận 9', '768'),
  (769, 79, 'Quận 10', '769'),
  (770, 79, 'Quận 11', '770'),
  (771, 79, 'Quận 12', '771'),
  (772, 79, 'Thủ Đức', '772');

-- ============================================
-- SAMPLE WARDS (Phường/Xã)
-- ============================================
-- Ba Dinh District Wards (Hanoi)
INSERT INTO wards
  (id, district_id, name, code)
VALUES
  (1, 1, 'Phúc Xá', '00001'),
  (2, 1, 'Trúc Bạch', '00002'),
  (3, 1, 'Vĩnh Phúc', '00003'),
  (4, 1, 'Cống Vị', '00004'),
  (5, 1, 'Liễu Giai', '00005'),
  (6, 1, 'Nguyễn Trung Trực', '00006'),
  (7, 1, 'Quán Thánh', '00007'),
  (8, 1, 'Ngọc Hà', '00008'),
  (9, 1, 'Điện Biên', '00009'),
  (10, 1, 'Đội Cấn', '00010');

-- Hoan Kiem District Wards (Hanoi)
INSERT INTO wards
  (id, district_id, name, code)
VALUES
  (100, 2, 'Phúc Tân', '00100'),
  (101, 2, 'Đồng Xuân', '00101'),
  (102, 2, 'Hàng Mã', '00102'),
  (103, 2, 'Hàng Buồm', '00103'),
  (104, 2, 'Hàng Đào', '00104'),
  (105, 2, 'Hàng Bồ', '00105'),
  (106, 2, 'Cửa Đông', '00106'),
  (107, 2, 'Lý Thái Tổ', '00107'),
  (108, 2, 'Hàng Bạc', '00108'),
  (109, 2, 'Hàng Gai', '00109');

-- District 1 Wards (Ho Chi Minh City)
INSERT INTO wards
  (id, district_id, name, code)
VALUES
  (26734, 760, 'Tân Định', '26734'),
  (26737, 760, 'Đa Kao', '26737'),
  (26740, 760, 'Bến Nghé', '26740'),
  (26743, 760, 'Bến Thành', '26743'),
  (26746, 760, 'Nguyễn Thái Bình', '26746'),
  (26749, 760, 'Phạm Ngũ Lão', '26749'),
  (26752, 760, 'Cầu Ông Lãnh', '26752'),
  (26755, 760, 'Cô Giang', '26755'),
  (26758, 760, 'Nguyễn Cư Trinh', '26758'),
  (26761, 760, 'Cầu Kho', '26761');

-- ============================================
-- NOTES
-- ============================================
-- This is a minimal seed with only major cities and sample districts/wards.
-- 
-- For complete Vietnam administrative divisions data, you should:
-- 1. Download complete data from: https://github.com/kenzouno1/DiaGioiHanhChinhVN
-- 2. Or use Vietnam General Statistics Office official data
-- 3. Or import from open datasets
--
-- Total count:
-- - 63 Cities/Provinces
-- - ~700+ Districts
-- - ~10,000+ Wards
--
-- To import complete data, use a script to generate SQL inserts or use COPY command:
-- COPY cities(id, name, code) FROM '/path/to/cities.csv' DELIMITER ',' CSV HEADER;

