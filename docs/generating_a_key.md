```bash
openssl genpkey -algorithm RSA -out scratch/keys/key.pem -pkeyopt rsa_keygen_bits:2048

openssl rsa -pubout -in scratch/keys/key.pem -out scratch/keys/key.pem.pub
```
