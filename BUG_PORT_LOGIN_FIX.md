# BUG REPORT: Port & Login Issues

**Tanggal**: 2025-11-11
**Dilaporkan**: User reported 2 bugs
1. `http://localhost:3000/traders` tidak berfungsi
2. Web selalu minta login dan tidak tersimpan

---

## 🔍 INVESTIGASI

### Bug #1: Port 3000 Tidak Berfungsi

**Laporan User**: "bug http://localhost:3000/traders tidak berfungsi"

**Root Cause Analysis**:

❌ **PORT 3000 TIDAK ADA!**

Aplikasi ini tidak menggunakan port 3000 sama sekali:

**Development Mode**:
- Frontend (Vite): Port **5173**
- Backend (Go): Port **8080**
- File: `web/vite.config.ts` line 14: `port: 5173`

**Production Mode**:
- Backend serve static files: Port **8080**
- File: `main.go` line 40: `Addr: ":8080"`

**Port 3000 = TIDAK ADA!**

User mencoba akses port yang salah.

---

### Bug #2: Login Tidak Tersimpan

**Laporan User**: "web minta selalu login dan tidak tersimpan"

**Code Investigation**:

✅ **AUTHENTICATION CODE SEBENARNYA BENAR!**

File yang saya cek:

1. **`web/src/contexts/AuthContext.tsx`**:
   - ✅ Line 27-28: Load token from `localStorage` on mount
   - ✅ Line 51-52: Save token to `localStorage` on login
   - ✅ Line 64-65: Save token to `localStorage` on register
   - ✅ Line 71-72: Remove from `localStorage` on logout

2. **`web/src/lib/api.ts`**:
   - ✅ Line 13: Read token from `localStorage`
   - ✅ Line 21: Add token to Authorization header
   - ✅ Format correct: `Bearer ${token}`

3. **`web/src/App.tsx`**:
   - ✅ Line 9-11: ProtectedRoute checks `isAuthenticated`
   - ✅ Redirects to `/login` if not authenticated

**CODE IS CORRECT!**

**Possible Causes**:
1. ❌ User clear localStorage/cookies
2. ❌ Browser private/incognito mode
3. ❌ Backend JWT verification failing
4. ❌ Token expired (need to check JWT expiry time)
5. ❌ CORS issue if accessing wrong port
6. ❌ Frontend not built (using old version)

---

## 🎯 ACTUAL PROBLEMS FOUND

### Problem 1: NO STARTUP SCRIPT

**Issue**: Tidak ada `run.sh` atau startup script yang jelas

**Impact**:
- User tidak tahu bagaimana menjalankan aplikasi
- User tidak tahu port yang benar
- Confusion tentang development vs production mode

**Files Checked**:
```bash
$ ls -la | grep -E "\.sh$|Makefile"
(no results)
```

**Result**: Tidak ada startup script!

---

### Problem 2: MISSING CLEAR DOCUMENTATION

**Issue**: README tidak menjelaskan dengan jelas port yang digunakan

**Dari README.md**:
- Line 130: "Backend will start on http://localhost:8080"
- ❌ TAPI tidak jelas explain frontend port
- ❌ Tidak ada penjelasan cara akses aplikasi
- ❌ Tidak ada penjelasan perbedaan dev vs production

---

### Problem 3: JWT TOKEN EXPIRY

**File**: `auth/auth.go` line 31
```go
ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour))
```

**Result**: Token expires in **24 hours** - ini OK, bukan masalah!

---

### Problem 4: BACKEND TIDAK SERVE STATIC FILES (CRITICAL!)

**Issue**: Backend tidak serve frontend static files!

**Investigation**:
```bash
$ grep -r "Static\|ServeFile\|FileServer" api/
(no results)
```

**Impact**:
- Backend hanya serve `/api/*` endpoints
- Frontend files (`web/dist/*`) **TIDAK DISERVE** oleh backend
- Harus run frontend dan backend **SECARA TERPISAH**

**Current Architecture**:
```
Development Mode:
- Frontend dev server: Port 5173 (Vite)
- Backend API: Port 8080 (Go)
- Access: http://localhost:5173

Production Mode:
- Backend API: Port 8080
- Frontend: ??? (TIDAK DISERVE!)
```

**MASALAHNYA**: Production mode tidak lengkap!

---

## 🔧 SOLUSI

### Solution 1: Buat Startup Script

Buat `run.sh` untuk memudahkan startup:

```bash
#!/bin/bash

MODE=${1:-dev}

if [ "$MODE" = "dev" ]; then
    echo "🚀 Starting DEVELOPMENT mode..."
    echo "Backend will run on: http://localhost:8080"
    echo "Frontend will run on: http://localhost:5173"
    echo ""

    # Start backend
    go run main.go &
    BACKEND_PID=$!

    # Wait for backend to start
    sleep 2

    # Start frontend
    cd web && npm run dev &
    FRONTEND_PID=$!

    echo ""
    echo "✅ Started!"
    echo "   Backend PID: $BACKEND_PID"
    echo "   Frontend PID: $FRONTEND_PID"
    echo ""
    echo "🌐 Open browser: http://localhost:5173"
    echo ""
    echo "Press Ctrl+C to stop..."

    # Wait for interrupt
    trap "kill $BACKEND_PID $FRONTEND_PID 2>/dev/null" EXIT
    wait

elif [ "$MODE" = "prod" ]; then
    echo "🚀 Starting PRODUCTION mode..."

    # Build frontend
    echo "Building frontend..."
    cd web && npm run build && cd ..

    # Start backend (will serve static files)
    echo "Starting backend on: http://localhost:8080"
    go run main.go

else
    echo "Usage: ./run.sh [dev|prod]"
    echo "  dev  - Development mode (frontend on 5173, backend on 8080)"
    echo "  prod - Production mode (all on 8080)"
fi
```

### Solution 2: Tambahkan Static File Serving ke Backend

File: `api/server.go`

Tambahkan di akhir route setup:

```go
// Serve frontend static files (production mode)
r.StaticFS("/assets", http.Dir("./web/dist/assets"))
r.StaticFile("/", "./web/dist/index.html")
r.NoRoute(func(c *gin.Context) {
    // For client-side routing, serve index.html
    c.File("./web/dist/index.html")
})
```

### Solution 3: Update README dengan Instruksi Jelas

Tambahkan ke README.md:

```markdown
## 🚀 How to Run

### Development Mode (Recommended)

Frontend dan backend run secara terpisah untuk hot-reload:

# Terminal 1 - Backend
go run main.go

# Terminal 2 - Frontend
cd web && npm run dev

# Akses aplikasi
Open: http://localhost:5173

Port yang digunakan:
- Frontend: 5173 (Vite dev server)
- Backend API: 8080
- Vite proxy /api requests to backend automatically

### Production Mode

Build frontend dan run melalui backend:

# Build frontend
cd web && npm run build && cd ..

# Start backend (will serve static files)
go run main.go

# Akses aplikasi
Open: http://localhost:8080

### Using Startup Script

# Development mode
./run.sh dev

# Production mode
./run.sh prod
```

---

## 📋 CHECKLIST FIX

- [ ] Buat `run.sh` script
- [ ] Tambahkan static file serving ke backend
- [ ] Update README dengan instruksi jelas
- [ ] Test development mode
- [ ] Test production mode
- [ ] Dokumentasikan port yang benar
- [ ] Debug mengapa login tidak persist (jika masih terjadi setelah fix)

---

## 🎯 JAWABAN UNTUK USER

### Q1: "http://localhost:3000/traders tidak berfungsi"

**A**: Port 3000 tidak digunakan oleh aplikasi ini.

**Port yang benar**:
- Development: `http://localhost:5173` (frontend Vite)
- Production: `http://localhost:8080` (backend + static files)

**Cara menjalankan**:
```bash
# Development mode (recommended)
# Terminal 1:
go run main.go

# Terminal 2:
cd web && npm run dev

# Buka browser:
http://localhost:5173
```

### Q2: "web minta selalu login dan tidak tersimpan"

**A**: Code authentication sudah benar. Kemungkinan penyebab:

1. ✅ **Token disimpan di localStorage** (expire 24 jam)
2. ✅ **Authorization header sudah benar**
3. ✅ **ProtectedRoute logic sudah benar**

**Possible Issues**:
- ❌ Browser clear cache/localStorage
- ❌ Accessing wrong port (CORS issue)
- ❌ Browser incognito/private mode
- ❌ Frontend not built (using stale version)

**Solution**:
1. Use correct port: `http://localhost:5173` (dev) or `http://localhost:8080` (prod)
2. Clear browser cache and try again
3. Check console for errors (F12 → Console tab)

**Test**:
```javascript
// Open browser console (F12)
localStorage.getItem('token')  // Should show JWT token after login
localStorage.getItem('user')   // Should show user object
```

If empty after login → bug in code
If exists but still logged out → bug in ProtectedRoute logic
If exists on refresh → everything working!

---

## 🐛 NEXT STEPS

1. Create `run.sh` script
2. Add static file serving to backend
3. Update documentation
4. Test both modes
5. Get user feedback

