# Sensor 06 — Load Tests (Gatling)

> **Objetivo:** validar que o serviço aguenta a carga esperada sem
> degradar latência ou aumentar taxa de erro.
> **Quando roda:** nightly (agendado) + manual antes de release
> de endpoints críticos.
> **Falha → ação:** **não bloqueia merge**; **bloqueia release**
> se regressão > 10% vs baseline.

---

## Ferramenta

**Gatling** — Scala/Java-based, mas com SDK JS/Java/Python.
Para Go, o backend precisa de uma build Gatling à parte (em
`test/load/`) usando o SDK Java ou JS.

> **Versão mínima:** Gatling 3.15.1+.
> **Enterprise opcional:** para integração com GitHub Actions
> centralizada, considerar Gatling Enterprise.

---

## Estrutura recomendada

```
test/load/
├── pom.xml                 # Maven (Java SDK)
├── src/test/java/
│   └── simulations/
│       ├── LoginSimulation.java
│       └── OrderFlowSimulation.java
├── src/test/resources/
│   ├── gatling.conf
│   └── bodies/
│       └── login.json
└── README.md
```

### Exemplo (Java SDK)

```java
package simulations;

import static io.gatling.javaapi.core.CoreDsl.*;
import static io.gatling.javaapi.http.HttpDsl.*;

import io.gatling.javaapi.core.*;
import io.gatling.javaapi.http.*;

public class LoginSimulation extends Simulation {
    HttpProtocolBuilder http = http
        .baseUrl("http://localhost:8080")
        .acceptHeader("application/json")
        .contentTypeHeader("application/json");

    ScenarioBuilder scn = scenario("Login")
        .exec(http("login")
            .post("/api/v1/auth/login")
            .body(StringBody("{\"email\":\"user@example.com\",\"password\":\"secret\"}"))
            .check(status().is(200))
            .check(jsonPath("$.token").saveAs("token")))
        .pause(2);

    {
        setUp(
            scn.injectOpen(
                rampUsers(100).during(30)
            ).protocols(http)
        ).assertions(
            global().successfulRequests().percent().gt(95.0),
            global().responseTime().percentile(95.0).lt(500)
        );
    }
}
```

---

## Comandos exatos

### Local (Maven)

```bash
cd test/load
mvn gatling:test -Dgatling.simulationClass=simulations.LoginSimulation
```

### Docker

```bash
docker run --rm -v $(pwd)/test/load:/opt/gatling gatlingio/gatling \
  -s simulations.LoginSimulation
```

### CI (com cache de relatório)

```yaml
- uses: actions/checkout@v4
- uses: actions/setup-java@v4
  with: { distribution: 'temurin', java-version: '17' }
- name: Run Gatling
  run: cd test/load && mvn gatling:test -Dgatling.simulationClass=simulations.LoginSimulation
- uses: actions/upload-artifact@v4
  with:
    name: gatling-report
    path: test/load/target/gatling/
```

---

## Thresholds (assertions)

| Métrica                            | Limite sugerido (ajustar por serviço) |
|------------------------------------|----------------------------------------|
| Taxa de sucesso (status 2xx)       | **≥ 95%**                              |
| Latência p95                       | **< 500ms**                            |
| Latência p99                       | **< 1500ms**                           |
| Throughput                         | **definir baseline**                   |
| Erro por segundo                   | **< 1**                                |

> **Baseline** é capturado no primeiro run e comparado em runs
> seguintes. Regressão > 10% em p95 ou taxa de erro → bloqueia release.

---

## Onde pluga no pipeline

### Nightly (`.github/workflows/nightly.yml`)

```yaml
load:
  name: Load tests (Gatling)
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-java@v4
      with: { distribution: 'temurin', java-version: '17' }
    - name: Bring up stack
      run: docker compose -f deploy/docker-compose.yml up -d
    - name: Wait for ready
      run: |
        timeout 60 bash -c 'until curl -fsS http://localhost:8080/healthz; do sleep 1; done'
    - name: Run Gatling
      run: |
        cd test/load
        mvn gatling:test \
          -Dgatling.simulationClass=simulations.LoginSimulation
    - name: Upload report
      if: always()
      uses: actions/upload-artifact@v4
      with:
        name: gatling-report
        path: test/load/target/gatling/
    - name: Teardown
      if: always()
      run: docker compose -f deploy/docker-compose.yml down
```

### Antes de release (manual / workflow_dispatch)

```yaml
on:
  workflow_dispatch:
    inputs:
      simulation:
        type: choice
        options: [LoginSimulation, OrderFlowSimulation, FullSuite]
```

---

## Quando **não** rodar

- Em todo PR (overhead, e baseline pode variar). Rodar **nightly** + antes
  de **release**.
- Em mudanças triviais (docs, configs). Apenas em mudanças de código
  de runtime.

---

## Falha típica & remediação

| Falha                                    | Como corrigir                                       |
|------------------------------------------|-----------------------------------------------------|
| p95 > 500ms (regressão)                  | Investigar: DB query lenta? GC? Falta de cache?     |
| Taxa de erro > 5%                        | Verificar logs do serviço durante o teste.         |
| Throughput muito abaixo do baseline      | Provável regressão; comparar com main.             |
| Timeout                                  | Aumentar timeout do Gatling ou investigar vazamento.|

---

## Quem roda

- **Nightly:** automático.
- **Antes de release:** `quality-assurance` (em coordenação com
  `devops-engineer`).
- **Falha:** registra issue `tech-debt` ou `bug`; **bloqueia release**
  se regressão significativa.
