# Windows Setup Guide

**Issue**: Bash script `run.sh` has line ending issues on Windows.

**Error**:
```
./run.sh: line 2: $'\r': command not found
./run.sh: line 37: syntax error near unexpected token `elif'
```

**Cause**: Windows uses CRLF line endings, bash expects LF.

---

## ✅ SOLUTION 1: Use Windows Batch Script (RECOMMENDED)

Gunakan file `run.bat` yang sudah disediakan:

### Development Mode:
```cmd
run.bat dev
```

### Production Mode:
```cmd
run.bat prod
```

---

## ✅ SOLUTION 2: Use PowerShell Script

Gunakan file `run.ps1`:

### Development Mode:
```powershell
.\run.ps1 dev
```

### Production Mode:
```powershell
.\run.ps1 prod
```

**Note**: Jika error "script tidak bisa dijalankan", jalankan:
```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

---

## ✅ SOLUTION 3: Manual Commands (PALING AMAN)

Jika script tidak berfungsi, jalankan manual:

### Development Mode:

**Terminal 1** (Backend):
```cmd
go run main.go
```

**Terminal 2** (Frontend):
```cmd
cd web
npm run dev
```

**Open Browser**:
```
http://localhost:5173
```

### Production Mode:

**Build Frontend**:
```cmd
cd web
npm run build
cd ..
```

**Start Backend**:
```cmd
go run main.go
```

**Open Browser**:
```
http://localhost:8080
```

---

## ✅ SOLUTION 4: Fix bash script (Git Bash/WSL)

Jika ingin pakai `run.sh` di Git Bash atau WSL:

### Fix Line Endings:
```bash
# Install dos2unix (if needed)
# Git Bash: Already included
# WSL: sudo apt install dos2unix

# Convert line endings
dos2unix run.sh

# Run script
./run.sh dev
```

### Or using Git:
```bash
# Configure git to auto-convert
git config core.autocrlf true

# Re-checkout file
rm run.sh
git checkout run.sh

# Run script
./run.sh dev
```

---

## 📋 PORT REFERENCE

Jangan lupa gunakan port yang benar:

| Mode | Port | URL | Description |
|------|------|-----|-------------|
| Development | 5173 | http://localhost:5173 | Frontend (Vite) |
| Development | 8080 | http://localhost:8080 | Backend API |
| Production | 8080 | http://localhost:8080 | Backend + Frontend |

**IMPORTANT**:
- ❌ Port 3000 **TIDAK DIGUNAKAN**
- ✅ Development: Buka `http://localhost:5173`
- ✅ Production: Buka `http://localhost:8080`

---

## 🐛 TROUBLESHOOTING

### Issue: "go: command not found"
**Solution**: Install Go dari https://go.dev/dl/

### Issue: "npm: command not found"
**Solution**: Install Node.js dari https://nodejs.org/

### Issue: Port already in use
**Solution**:
```cmd
# Check what's using the port
netstat -ano | findstr :8080
netstat -ano | findstr :5173

# Kill the process (replace PID)
taskkill /PID <PID> /F
```

### Issue: Login tidak tersimpan
**Solution**:
1. Pastikan pakai port yang benar (5173 dev, 8080 prod)
2. Clear browser cache: Ctrl+Shift+Delete
3. Check localStorage di Console (F12):
   ```javascript
   localStorage.getItem('token')
   localStorage.getItem('user')
   ```

### Issue: CORS error
**Solution**:
- Pastikan backend running di port 8080
- Pastikan frontend di port 5173 (dev mode)
- Backend sudah dikonfigurasi CORS untuk kedua port

---

## 🎯 QUICK START (Windows)

Cara tercepat untuk mulai:

1. **Open Command Prompt** (cmd)

2. **Navigate to project**:
   ```cmd
   cd C:\Users\ajulr\LLM-Trend
   ```

3. **Run batch script**:
   ```cmd
   run.bat dev
   ```

4. **Wait for startup** (2-3 seconds)

5. **Open browser**:
   ```
   http://localhost:5173
   ```

6. **Done!** ✅

---

## 📝 ADDITIONAL NOTES

### File Locations:
- `run.bat` - Windows batch script (RECOMMENDED)
- `run.ps1` - PowerShell script (alternative)
- `run.sh` - Bash script (for Git Bash/WSL)

### Which Script to Use?
- **Windows CMD**: Use `run.bat` ✅
- **PowerShell**: Use `run.ps1`
- **Git Bash**: Fix `run.sh` first with `dos2unix`
- **WSL**: Fix `run.sh` first with `dos2unix`
- **Manual**: Run commands directly (safest)

### Performance Tips:
- Close other applications using ports 8080/5173
- Use production mode for better performance
- Development mode has hot-reload (slower but convenient)

---

## ✅ RECOMMENDED FOR YOU

Based on your setup (Windows), gunakan:

```cmd
run.bat dev
```

Kemudian buka browser:
```
http://localhost:5173
```

**Jangan pakai port 3000** - port itu tidak digunakan!
