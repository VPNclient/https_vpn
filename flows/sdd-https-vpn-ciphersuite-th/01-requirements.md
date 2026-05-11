# ข้อกำหนด: sdd-https-vpn-ciphersuite-th (ชุดรหัสลับสำหรับประเทศไทย)

> เวอร์ชัน: 1.2  
> สถานะ: ร่าง (DRAFT)  
> อัปเดตล่าสุด: 2026-05-11

## ปัญหาที่ต้องการแก้ไข (Problem Statement)

เพื่อเสริมสร้างความปลอดภัยของ HTTPS VPN ให้รองรับมาตรฐานการเข้ารหัสลับสมัยใหม่ของประเทศไทย และเตรียมพร้อมสำหรับการเปลี่ยนแปลงสู่ยุค Quantum-Resistant Cryptography (QRC) ตามคำแนะนำของ NIST (FIPS 203, 204, 205) โดยสแต็กการเข้ารหัสนี้จะถูกออกแบบให้ทำงานร่วมกับสถาปัตยกรรม HTTPS VPN เดิมได้อย่างไร้รอยต่อ (Seamless Integration) เพื่อรักษาความสมดุลระหว่างประสิทธิภาพและความปลอดภัยสูงสุด

## เรื่องราวของผู้ใช้ (User Stories)

### เรื่องราวหลัก (Primary)

**ในฐานะ** ผู้ดูแลระบบความปลอดภัย (Security Administrator)  
**ฉันต้องการ** ใช้ระบบการแลกเปลี่ยนกุญแจแบบไฮบริด (Hybrid KEM) และการลงนามดิจิทัลที่เหมาะสมกับสถาปัตยกรรม TLS-based ของ HTTPS VPN  
**เพื่อให้** ข้อมูลการสื่อสารมีความมั่นคงปลอดภัยสูงสุดในขณะที่ยังคงประสิทธิภาพสำหรับทราฟฟิกทั่วไป

## เกณฑ์การยอมรับ (Acceptance Criteria)

### สิ่งที่ต้องมี (Must Have)

1. **การแลกเปลี่ยนกุญแจ (Key Exchange):**
   - **HTTPS VPN Compatibility:** ระบบต้องสามารถแทรก Hybrid KEM เข้าไปในขั้นตอน TLS Handshake ของ HTTPS VPN ได้โดยไม่กระทบต่อโปรโตคอลรับส่งข้อมูลหลัก
   - **Balanced Profile (Default):** ต้องใช้ Hybrid KEM ระหว่าง X25519 หรือ P-256 ร่วมกับ ML-KEM-768 สำหรับ Control Plane และ Session Establishment ทั่วไป
   - **High-Assurance Profile:** ต้องรองรับ ML-KEM-1024 สำหรับช่องทางผู้ดูแลระบบ (Admin Channels), การลงทะเบียน (Enrollment), และการตั้งค่าช่องทางระยะยาว (Long-term setup)

2. **การลงนามดิจิทัล (Digital Signatures):**
   - **Operational Use:** ใช้ Hybrid Ed25519/ECDSA + ML-DSA-65 สำหรับใบรับรองทั่วไป, Control API signing และ Signed config bundles
   - **Conservative Trust Anchor:** ใช้ SLH-DSA สำหรับ Root Manifest (Offline), Firmware Signing fallback และ Disaster Recovery keys

3. **การลงนามเฟิร์มแวร์ (Firmware Signing):**
   - **Integrity Protection:** ต้องมีเครื่องมือสำหรับลงนามไฟล์เฟิร์มแวร์ (`h2_firmware`) โดยใช้ SLH-DSA เพื่อป้องกันการแก้ไขจากคอมพิวเตอร์ควอนตัม
   - **Multi-layer Signature:** รองรับการลงนามซ้อน (Double signing) ระหว่าง Ed25519 และ ML-DSA-65 สำหรับการตรวจสอบในหลายระดับ

4. **ระบบสำรอง (Backup):**
   - ต้องมี HQC (Hamming Quasi-Cyclic) เป็น Backup KEM ในกรณีที่ ML-KEM พบช่องโหว่

### สิ่งที่ควรมี (Should Have)

- การเตรียมความพร้อมสำหรับมาตรฐานที่ใช้ Falcon ในอนาคตตามทิศทางของ NIST
- การวิเคราะห์ช่องว่าง (Gap Analysis) สำหรับการเปลี่ยนผ่านจาก Classical เป็น PQC ทั้งระบบ

## ข้อจำกัด (Constraints)

- **Performance**: ML-KEM-1024 จะไม่ถูกใช้เป็นค่าเริ่มต้นสำหรับทราฟฟิกที่ไวต่อความหน่วง (Latency-sensitive) เนื่องจากขนาดกุญแจและภาระการประมวลผล
- **Complexity**: การจัดการใบรับรองแบบ Hybrid จะมีความซับซ้อนเพิ่มขึ้นในการจัดการ Trust Chain

## คำถามที่ยังไม่มีคำตอบ (Open Questions)

- [ ] การจัดการขนาด MTU ของ VPN เมื่อใช้ SLH-DSA ซึ่งมีขนาดลายเซ็นที่ใหญ่มาก
- [ ] ความพร้อมของ Library สำหรับ ML-DSA และ SLH-DSA ในเวอร์ชัน Production

## เอกสารอ้างอิง (References)

- FIPS 203: ML-KEM (Module-Lattice-Based Key-Encapsulation Mechanism)
- FIPS 204: ML-DSA (Module-Lattice-Based Digital Signature Standard)
- FIPS 205: SLH-DSA (Stateless Hash-Based Digital Signature Standard)
- NIST PQC Project: HQC, Falcon (ongoing research)

---

## การอนุมัติ (Approval)

- [ ] ตรวจสอบโดย: [ชื่อ]
- [ ] อนุมัติเมื่อ: [วันที่]
- [ ] หมายเหตุ: [ข้อเสนอแนะเพิ่มเติม]

## ข้อจำกัด (Constraints)

- **Performance**: ML-KEM-1024 จะไม่ถูกใช้เป็นค่าเริ่มต้นสำหรับทราฟฟิกที่ไวต่อความหน่วง (Latency-sensitive) เนื่องจากขนาดกุญแจและภาระการประมวลผล
- **Complexity**: การจัดการใบรับรองแบบ Hybrid จะมีความซับซ้อนเพิ่มขึ้นในการจัดการ Trust Chain

## คำถามที่ยังไม่มีคำตอบ (Open Questions)

- [ ] การจัดการขนาด MTU ของ VPN เมื่อใช้ SLH-DSA ซึ่งมีขนาดลายเซ็นที่ใหญ่มาก
- [ ] ความพร้อมของ Library สำหรับ ML-DSA และ SLH-DSA ในเวอร์ชัน Production

## เอกสารอ้างอิง (References)

- FIPS 203: ML-KEM (Module-Lattice-Based Key-Encapsulation Mechanism)
- FIPS 204: ML-DSA (Module-Lattice-Based Digital Signature Standard)
- FIPS 205: SLH-DSA (Stateless Hash-Based Digital Signature Standard)
- NIST PQC Project: HQC, Falcon (ongoing research)

---

## การอนุมัติ (Approval)

- [ ] ตรวจสอบโดย: [ชื่อ]
- [ ] อนุมัติเมื่อ: [วันที่]
- [ ] หมายเหตุ: [ข้อเสนอแนะเพิ่มเติม]
