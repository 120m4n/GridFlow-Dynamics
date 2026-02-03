# Test - Simulador de Cuadrilla

Este directorio contiene scripts de prueba para simular mensajes de inventario de cuadrilla hacia la API de GridFlow-Dynamics.

## Simulador Python

`simulate_cuadrilla.py` - Script funcional que genera y envía mensajes aleatorios con firma HMAC-SHA256.

### Características

- ✅ **Programación Funcional**: Usa funciones puras, composición y higher-order functions
- ✅ **Datos Aleatorios**: Genera coordenadas, estados y progreso aleatorios
- ✅ **Firma HMAC-SHA256**: Calcula automáticamente la firma de cada mensaje
- ✅ **Delays Aleatorios**: Espaciado de 1-57 segundos entre peticiones
- ✅ **Timestamps Locales**: Usa hora del sistema en formato ISO8601
- ✅ **Logging Completo**: Muestra detalles de cada petición y respuesta

### Configuración de Coordenadas

El script usa rangos específicos de coordenadas (región Barranquilla, Colombia):

- **Longitud**: -74.89612235727202 a -74.81308123746624
- **Latitud**: 10.9450893430744 a 11.04217720876936

### Instalación de Dependencias

```bash
pip install requests
```

### Uso

**Modo continuo** (detener con Ctrl+C):
```bash
python test/simulate_cuadrilla.py
```

**Modo limitado** (número específico de peticiones):
```bash
# Enviar 10 peticiones
python test/simulate_cuadrilla.py 10

# Enviar 50 peticiones
python test/simulate_cuadrilla.py 50
```

### Personalización

Edita las siguientes constantes en el script para personalizar:

```python
# URL del API
API_URL = "http://localhost:8080/api/v1/mensaje_inventario/cuadrilla"

# Secreto HMAC (debe coincidir con el servidor)
HMAC_SECRET = "default-secret-change-in-production"

# Rango de delay entre requests
DELAY_RANGE = (1, 57)

# Cuadrillas y empleados simulados
CUADRILLAS = ["G0/CUADRILLA_001", "G0/CUADRILLA_002", ...]
EMPLEADOS = ["Juan Perez", "Maria Rodriguez", ...]
```

### Ejemplo de Salida

```
🚀 Iniciando simulación de mensajes de cuadrilla
📍 API URL: http://localhost:8080/api/v1/mensaje_inventario/cuadrilla
⏱️  Delay entre peticiones: 1-57 segundos
🔢 Peticiones: infinitas

================================================================================
[2026-02-03 14:23:45] Enviando petición...
Cuadrilla: G0/CUADRILLA_002
Empleado: Maria Rodriguez
Estado: trabajando
Coordenadas: (10.987654, -74.876543)
Progreso: 67%
Batería: 78%
Firma HMAC: a3f5d8e9b2c4f6a1...

✅ Respuesta: 200
Body: {'status': 'success', 'message': 'Mensaje de inventario de cuadrilla recibido correctamente.'}
================================================================================

⏳ Esperando 23 segundos antes de la siguiente petición...
```

### Arquitectura Funcional

El script sigue principios de programación funcional:

1. **Funciones Puras**: Generación de datos sin efectos secundarios
2. **Composición**: `create_signature = calculate_hmac ∘ serialize_payload`
3. **Higher-Order Functions**: Generadores de valores aleatorios
4. **Separación de Efectos**: I/O aislado en funciones específicas

```python
# Funciones puras (sin efectos secundarios)
generate_random_payload() -> Dict
calculate_hmac_signature(data, secret) -> str
create_headers(signature) -> Dict

# Efectos I/O (separados)
send_request(url, payload, headers) -> Response
log_request(payload, signature) -> None
```
