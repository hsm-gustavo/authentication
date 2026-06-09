 ALTER TABLE "recoveries" ADD COLUMN "type" VARCHAR(20) NOT NULL DEFAULT 'password_recovery';

-- tipos são: 'email_verification' ou 'password_recovery' 