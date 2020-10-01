-- --------------------------------------------------------
-- Host:                         127.0.0.1
-- Server version:               10.4.12-MariaDB - mariadb.org binary distribution
-- Server OS:                    Win64
-- HeidiSQL Version:             10.3.0.5771
-- --------------------------------------------------------

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET NAMES utf8 */;
/*!50503 SET NAMES utf8mb4 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;


-- Dumping database structure for go_exercies
DROP DATABASE IF EXISTS `go_exercies`;
CREATE DATABASE IF NOT EXISTS `go_exercies` /*!40100 DEFAULT CHARACTER SET utf8 */;
USE `go_exercies`;

-- Dumping structure for table go_exercies.role
DROP TABLE IF EXISTS `role`;
CREATE TABLE IF NOT EXISTS `role` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(64) NOT NULL,
  `description` varchar(128) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8;

-- Dumping data for table go_exercies.role: ~2 rows (approximately)
DELETE FROM `role`;
/*!40000 ALTER TABLE `role` DISABLE KEYS */;
INSERT INTO `role` (`id`, `name`, `description`) VALUES
	(1, 'ADMIN', 'Super user'),
	(2, 'KASIR', 'User penjual'),
	(3, 'INVENTORI', 'User pemantau Stok');
/*!40000 ALTER TABLE `role` ENABLE KEYS */;

-- Dumping structure for table go_exercies.user
DROP TABLE IF EXISTS `user`;
CREATE TABLE IF NOT EXISTS `user` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `role_id` int(10) unsigned NOT NULL,
  `username` varchar(32) NOT NULL,
  `password` varchar(255) NOT NULL,
  `active` tinyint(4) NOT NULL DEFAULT 1,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `username` (`username`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8;

-- Dumping data for table go_exercies.user: ~0 rows (approximately)
DELETE FROM `user`;
/*!40000 ALTER TABLE `user` DISABLE KEYS */;
INSERT INTO `user` (`id`, `role_id`, `username`, `password`, `active`, `created_at`, `updated_at`, `deleted_at`) VALUES
	(1, 1, 'admin', '$2a$10$WNLtogj3RX9YuNBpj7rnJe.4LYNNJdBPrrL34JtInpH0LZRMVPzb.', 1, '2020-10-01 09:10:19', '2020-10-01 09:10:27', NULL),
	(2, 2, 'test', '$2a$10$aWSxj4B6gJHbZ24LWUBdD.R06kDTOsrZ6Go0R/LIWRbXHkSP4pMty', 1, '2020-10-01 11:24:13', '2020-10-01 17:41:56', NULL),
	(3, 2, 'kasir2', '$2a$10$nH/0UCdaUr3CaK3BHY/OeOz8GRU8lv7Iw78SAcEZcScS08kbjeChi', 0, '2020-10-01 17:39:40', '2020-10-01 17:40:41', '2020-10-01 17:41:14'),
	(4, 3, 'kasir', '$2a$10$nw2MATPUOts4B9rzOh9B.eKkVtt89IPTlrDkCnbgjfLjgoZUX8JNi', 1, '2020-10-01 17:58:49', '2020-10-01 17:58:49', NULL);
/*!40000 ALTER TABLE `user` ENABLE KEYS */;

/*!40101 SET SQL_MODE=IFNULL(@OLD_SQL_MODE, '') */;
/*!40014 SET FOREIGN_KEY_CHECKS=IF(@OLD_FOREIGN_KEY_CHECKS IS NULL, 1, @OLD_FOREIGN_KEY_CHECKS) */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
