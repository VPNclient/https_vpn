# ระบบรหัสลับหลังควอนตัมสำหรับประเทศไทย (Thai Post-Quantum Ciphersuite - TH-PQC)

> [English Version](README_en.md) | **ภาษาไทย**

โปรเจกต์นี้เป็นการนำมาตรฐานการเข้ารหัสลับหลังควอนตัม (Post-Quantum Cryptography) มาปรับใช้ในสถาปัตยกรรม HTTPS VPN สำหรับประเทศไทย เพื่อเตรียมความพร้อมต่อภัยคุกคามจากคอมพิวเตอร์ควอนตัม โดยอ้างอิงมาตรฐาน NIST FIPS 203, 204 และ 205

## ฟีเจอร์หลัก (Key Features)

### 1. การแลกเปลี่ยนกุญแจแบบไฮบริด (Hybrid KEM)
รองรับการแลกเปลี่ยนกุญแจที่ผสมผสานระหว่างอัลกอริทึมคลาสสิกและ PQC:
- **Balanced Profile (ค่าเริ่มต้น):** X25519 + ML-KEM-768 สำหรับทราฟฟิกทั่วไป
- **High-Assurance Profile:** P-384 + ML-KEM-1024 สำหรับช่องทางผู้ดูแลระบบ

### 2. ระบบการลงนามดิจิทัล (Digital Signatures)
- **Operational Use:** Hybrid Ed25519/ECDSA + ML-DSA-65 สำหรับใบรับรองและ API
- **Conservative Trust Anchor:** SLH-DSA (FIPS 205) สำหรับ Root Manifest และการลงนามเฟิร์มแวร์

### 3. ระบบสำรองและความปลอดภัย (Security & Backup)
- **HQC (Backup KEM):** ระบบสำรองในกรณีที่อัลกอริทึมแบบ Lattice ถูกโจมตี
- **Hybrid Logic:** ใช้การรวมกุญแจแบบ HKDF เพื่อความปลอดภัยสูงสุด

## โครงสร้างเอกสาร SDD (Documentation)

- [01-ข้อกำหนด (Requirements)](01-requirements.md)
- [02-รายละเอียดทางเทคนิค (Specifications)](02-specifications.md)
- [03-แผนการดำเนินงาน (Plan)](03-plan.md)
- [04-บันทึกการดำเนินการ (Implementation Log)](04-implementation-log.md)

## สถานะโครงการ (Project Status)

✅ **เสร็จสมบูรณ์ (COMPLETED)** - อิมพลีเมนต์โครงสร้างพื้นฐานและผ่านการทดสอบ Unit Test ทั้งหมดแล้ว

---
พัฒนาโดย Gemini CLI สำหรับโครงการ HTTPS VPN
