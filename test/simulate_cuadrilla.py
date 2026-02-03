#!/usr/bin/env python3
"""
Simulador de mensajes de inventario de cuadrilla usando programación funcional.
Envía solicitudes POST al endpoint de GridFlow-Dynamics con firma HMAC-SHA256.
"""

import hmac
import hashlib
import json
import random
import time
from datetime import datetime
from typing import Dict, Callable, Any
from functools import partial
import requests


# ============================================================================
# CONFIGURACIÓN
# ============================================================================

API_URL = "http://localhost:8080/api/v1/mensaje_inventario/cuadrilla"
HMAC_SECRET = "your-secret-key"

# Rangos de coordenadas
LONGITUDE_RANGE = (-74.89612235727202, -74.81308123746624)
LATITUDE_RANGE = (10.9450893430744, 11.04217720876936)

# Rango de delay entre requests (segundos)
DELAY_RANGE = (1, 57)

# Estados posibles
ESTADOS = ["en_ruta", "trabajando", "en_pausa", "finalizado"]

# Cuadrillas disponibles
CUADRILLAS = [
    "G0/CUADRILLA_001",
    "G0/CUADRILLA_002",
    "G0/CUADRILLA_003",
    "G1/CUADRILLA_101",
    "G1/CUADRILLA_102",
]

# Empleados disponibles
EMPLEADOS = [
    "Juan Perez",
    "Maria Rodriguez",
    "Carlos Gomez",
    "Ana Martinez",
    "Luis Fernandez",
]


# ============================================================================
# FUNCIONES PURAS - Generación de datos aleatorios
# ============================================================================

def random_float(min_val: float, max_val: float) -> Callable[[], float]:
    """Retorna función que genera float aleatorio en el rango dado."""
    return lambda: random.uniform(min_val, max_val)


def random_choice(options: list) -> Callable[[], Any]:
    """Retorna función que selecciona elemento aleatorio de una lista."""
    return lambda: random.choice(options)


def random_int(min_val: int, max_val: int) -> Callable[[], int]:
    """Retorna función que genera entero aleatorio en el rango dado."""
    return lambda: random.randint(min_val, max_val)


def current_timestamp() -> str:
    """Retorna timestamp actual en formato ISO8601."""
    return datetime.now().astimezone().isoformat()


# ============================================================================
# FUNCIONES PURAS - Construcción del payload
# ============================================================================

def create_coordenadas(lat_fn: Callable, lon_fn: Callable) -> Dict[str, float]:
    """Crea diccionario de coordenadas usando funciones generadoras."""
    return {
        "latitud": lat_fn(),
        "longitud": lon_fn()
    }


def create_payload(
    grupo_trabajo: str,
    nombre_empleado: str,
    timestamp: str,
    coordenadas: Dict[str, float],
    codigo_odt: str,
    estado: str,
    porcentaje_progreso: int,
    nivel_bateria: int
) -> Dict[str, Any]:
    """Construye el payload del mensaje de inventario."""
    return {
        "grupoTrabajo": grupo_trabajo,
        "nombreEmpleado": nombre_empleado,
        "timestamp": timestamp,
        "coordenadas": coordenadas,
        "codigoODT": codigo_odt,
        "estado": estado,
        "procentajeProgreso": porcentaje_progreso,
        "nivelBateria": nivel_bateria
    }


def generate_random_payload() -> Dict[str, Any]:
    """Genera payload aleatorio usando composición de funciones."""
    # Generadores de valores aleatorios
    lat_gen = random_float(*LATITUDE_RANGE)
    lon_gen = random_float(*LONGITUDE_RANGE)
    cuadrilla_gen = random_choice(CUADRILLAS)
    empleado_gen = random_choice(EMPLEADOS)
    estado_gen = random_choice(ESTADOS)
    progreso_gen = random_int(0, 100)
    bateria_gen = random_int(20, 100)
    
    # Generador de ODT consecutivo
    odt_gen = lambda: f"ODT_{random.randint(1000, 9999)}_{random.randint(100, 999)}"
    
    return create_payload(
        grupo_trabajo=cuadrilla_gen(),
        nombre_empleado=empleado_gen(),
        timestamp=current_timestamp(),
        coordenadas=create_coordenadas(lat_gen, lon_gen),
        codigo_odt=odt_gen(),
        estado=estado_gen(),
        porcentaje_progreso=progreso_gen(),
        nivel_bateria=bateria_gen()
    )


# ============================================================================
# FUNCIONES PURAS - Firma HMAC
# ============================================================================

def serialize_payload(payload: Dict[str, Any]) -> bytes:
    """Serializa el payload a JSON bytes (exactamente como se envía)."""
    # Importante: debe ser exactamente el mismo formato que requests.post enviará
    return json.dumps(payload, ensure_ascii=False).encode('utf-8')


def calculate_hmac_signature(data: bytes, secret: str) -> str:
    """Calcula la firma HMAC-SHA256 del data con el secreto dado."""
    return hmac.new(
        secret.encode('utf-8'),
        data,
        hashlib.sha256
    ).hexdigest()


def create_signature_and_body(payload: Dict[str, Any], secret: str) -> tuple[str, bytes]:
    """
    Composición: serializa payload y calcula firma HMAC.
    Retorna tupla (signature, body_bytes) para asegurar consistencia.
    """
    body_bytes = serialize_payload(payload)
    signature = calculate_hmac_signature(body_bytes, secret)
    return signature, body_bytes


# ============================================================================
# FUNCIONES PURAS - Headers
# ============================================================================

def create_headers(signature: str) -> Dict[str, str]:
    """Crea los headers HTTP necesarios para la petición."""
    return {
        "Content-Type": "application/json",
        "X-Signature-256": signature
    }


# ============================================================================
# EFECTOS - Operaciones I/O
# ============================================================================

def send_request(
    url: str,
    body_bytes: bytes,
    headers: Dict[str, str]
) -> requests.Response:
    """Envía la petición POST al API con el body exacto usado para HMAC."""
    return requests.post(
        url,
        data=body_bytes,
        headers=headers,
        timeout=10
    )


def log_request(payload: Dict[str, Any], signature: str) -> None:
    """Registra información de la petición."""
    print(f"\n{'='*80}")
    print(f"[{datetime.now().strftime('%Y-%m-%d %H:%M:%S')}] Enviando petición...")
    print(f"Cuadrilla: {payload['grupoTrabajo']}")
    print(f"Empleado: {payload['nombreEmpleado']}")
    print(f"Estado: {payload['estado']}")
    print(f"Coordenadas: ({payload['coordenadas']['latitud']:.6f}, "
          f"{payload['coordenadas']['longitud']:.6f})")
    print(f"Progreso: {payload['procentajeProgreso']}%")
    print(f"Batería: {payload['nivelBateria']}%")
    print(f"Firma HMAC: {signature[:16]}...")


def log_response(response: requests.Response) -> None:
    """Registra información de la respuesta."""
    status_emoji = "✅" if response.status_code == 200 else "❌"
    print(f"\n{status_emoji} Respuesta: {response.status_code}")
    try:
        print(f"Body: {response.json()}")
    except Exception:
        print(f"Body: {response.text}")
    print(f"{'='*80}\n")


def sleep_random_interval(min_seconds: int, max_seconds: int) -> None:
    """Pausa la ejecución por un intervalo aleatorio."""
    delay = random.randint(min_seconds, max_seconds)
    print(f"⏳ Esperando {delay} segundos antes de la siguiente petición...")
    time.sleep(delay)


# ============================================================================
# FUNCIÓN PRINCIPAL - Composición de efectos
# ============================================================================

def send_cuadrilla_message(url: str, secret: str) -> None:
    """
    Función principal que compone todas las operaciones:
    1. Genera payload aleatorio
    2. Calcula firma HMAC sobre los bytes exactos
    3. Crea headers
    4. Envía petición con los mismos bytes
    5. Registra resultados
    """
    try:
        # Generar datos (puro)
        payload = generate_random_payload()
        signature, body_bytes = create_signature_and_body(payload, secret)
        headers = create_headers(signature)
        
        # Efectos I/O
        log_request(payload, signature)
        response = send_request(url, body_bytes, headers)
        log_response(response)
        
    except requests.exceptions.RequestException as e:
        print(f"❌ Error en la petición: {e}")
    except Exception as e:
        print(f"❌ Error inesperado: {e}")


def run_simulation(
    url: str,
    secret: str,
    num_requests: int = None,
    delay_range: tuple = DELAY_RANGE
) -> None:
    """
    Ejecuta la simulación enviando múltiples peticiones.
    
    Args:
        url: URL del endpoint API
        secret: Secreto HMAC
        num_requests: Número de peticiones (None = infinito)
        delay_range: Tupla (min, max) para delay aleatorio entre peticiones
    """
    print("🚀 Iniciando simulación de mensajes de cuadrilla")
    print(f"📍 API URL: {url}")
    print(f"⏱️  Delay entre peticiones: {delay_range[0]}-{delay_range[1]} segundos")
    print(f"🔢 Peticiones: {'infinitas' if num_requests is None else num_requests}")
    
    try:
        count = 0
        while num_requests is None or count < num_requests:
            send_cuadrilla_message(url, secret)
            count += 1
            
            if num_requests is None or count < num_requests:
                sleep_random_interval(*delay_range)
                
    except KeyboardInterrupt:
        print(f"\n\n⛔ Simulación detenida por el usuario")
        print(f"📊 Total de peticiones enviadas: {count}")


# ============================================================================
# ENTRY POINT
# ============================================================================

if __name__ == "__main__":
    import sys
    
    # Verificar argumentos de línea de comandos
    if len(sys.argv) > 1:
        try:
            num_requests = int(sys.argv[1])
            print(f"Modo: Enviar {num_requests} peticiones")
            run_simulation(API_URL, HMAC_SECRET, num_requests=num_requests)
        except ValueError:
            print("❌ Error: El argumento debe ser un número entero")
            print("Uso: python simulate_cuadrilla.py [numero_de_peticiones]")
            sys.exit(1)
    else:
        print("Modo: Peticiones continuas (Ctrl+C para detener)")
        run_simulation(API_URL, HMAC_SECRET)
