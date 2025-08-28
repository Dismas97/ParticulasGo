package main

import (
		"encoding/json"
		"errors"
		"io"
		"math"
		"math/rand"
		"os"
		"runtime"
		"strings"
		"github.com/go-gl/gl/v4.1-core/gl"
		"github.com/go-gl/glfw/v3.3/glfw"
		"github.com/go-gl/mathgl/mgl32"
		log "github.com/sirupsen/logrus"
		"gopkg.in/natefinch/lumberjack.v2"
)

type Config struct {
		OpenglVerMay int `json:"opengl_v_mayor"`
		OpenglVerMen int `json:"opengl_v_menor"`
		Redimensionable bool `json:"redimensionable"`
		Ancho int `json:"ancho"`
		Alto int `json:"alto"`		
		Particulas int `json:"particulas"`
		Fuerza float64  `json:"fuerza_max"`
}

type Particula struct {
		PX float32
		PY float32
		ColorR float32
		ColorG float32
		ColorB float32
		VX float32
		VY float32
}

var (
		prevX, prevY   int
		prevWidth      int
		prevHeight     int
		isFullscreen   bool
		monitor *glfw.Monitor
)

var vertexShaderSrc =
`#version 410 core
layout (location = 0) in vec2 aPos;
layout (location = 1) in vec3 aColor;

out vec3 ourColor;
uniform mat4 projection;

void main() {
    gl_Position = projection * vec4(aPos, 0.0, 1.0);
    ourColor = aColor;
}`

var fragmentShaderSrc =
`#version 410 core
in vec3 ourColor;
out vec4 FragColor;

void main() {
    FragColor = vec4(ourColor, 1.0);
}`

var particulas = []Particula{
		{PX: 100, PY: 100, ColorR: 1, ColorG: 1, ColorB: 1, VX: 0, VY: 0},
}

func genParticulas(cantidad int) {
		var particulaAux Particula
		for i := 0; i < cantidad; i++ {				
				aux := rand.Float32() * 1920
				for aux == 0 {								
						aux = rand.Float32() * 1920
				}
				particulaAux.PX = aux
				
				aux = rand.Float32() * 1080
				for aux == 0 {								
						aux = rand.Float32() * 1080
				}
				
				particulaAux.PY = aux

				aux = rand.Float32()
				for aux == 0 {								
						aux = rand.Float32()
				}

				particulaAux.ColorR = aux

				aux = rand.Float32()
				for aux == 0 {								
						aux = rand.Float32()
				}
				
				particulaAux.ColorG = aux

				aux = rand.Float32()
				for aux == 0 {								
						aux = rand.Float32()
				}
				
				particulaAux.ColorB = aux

				particulas = append(particulas, particulaAux)
		}
}


func init() {
		runtime.LockOSThread()
}

func compilarShader(s string, stipo uint32) (uint32, error) {
		shader := gl.CreateShader(stipo)
		csources, free := gl.Strs(s)
		gl.ShaderSource(shader, 1, csources, nil)
		free()
		gl.CompileShader(shader)
    
		var status int32
		gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
		if status == gl.FALSE {
				var logLength int32
				gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)
				logs := strings.Repeat("\x00", int(logLength+1))
				gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(logs))
				return 0, errors.New("Error al compilar shader:"+ logs)
		}
		return shader, nil
}

func main() {
		logRotator := &lumberjack.Logger{
				Filename:   "logs.log",
				MaxSize:    10,         //MB
				MaxBackups: 3,
				MaxAge:     28,         // Días
				Compress:   true,
		}
		
		multiWriter := io.MultiWriter(os.Stdout, logRotator)
		log.SetOutput(multiWriter)
		log.SetLevel(log.DebugLevel)
		log.SetReportCaller(true)
		file, _ := os.Open("config.json")
		defer file.Close()

		var config Config
		decoder := json.NewDecoder(file)
		_ = decoder.Decode(&config)
		log.Debug(config)
		
		if err := glfw.Init(); err != nil {
				log.Fatalln("Ups: ", err)
		}
		defer glfw.Terminate()

		log.Debug("Todo bien")
		glfw.WindowHint(glfw.ContextVersionMajor, config.OpenglVerMay)
		glfw.WindowHint(glfw.ContextVersionMinor, config.OpenglVerMen)
		glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
		glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
		glfw.WindowHint(glfw.Resizable, glfw.True)

		prevHeight = config.Alto
		prevWidth = config.Ancho
		
		window, err := glfw.CreateWindow(config.Ancho, config.Alto, "Triángulo OpenGL (F11 = pantalla completa)", nil, nil)
		if err != nil {
				log.Fatalln("Error al crear ventana:", err)
		}
		window.MakeContextCurrent()

		if err := gl.Init(); err != nil {
				log.Fatalln("Erro al inicializar bindings opengl:", err)
		}

		version := gl.GoStr(gl.GetString(gl.VERSION))
		log.Debugf("Version OpenGL: %s", version)

		monitor = glfw.GetPrimaryMonitor()

		genParticulas(config.Particulas)
		
		var vertices []float32
		for _, p := range particulas {
				vertices = append(vertices, p.PX, p.PY, p.ColorR, p.ColorG, p.ColorB)
		}


		var vao, vbo uint32
		gl.GenVertexArrays(1, &vao)
		gl.BindVertexArray(vao)

		gl.GenBuffers(1, &vbo)
		gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
		gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

		// Posición, tamaño, tipo, normalizado, salto, puntero
		gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 5*4, gl.PtrOffset(0))
		gl.EnableVertexAttribArray(0)

		gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 5*4, gl.PtrOffset(2*4))
		gl.EnableVertexAttribArray(1)

		vertexShader, err := compilarShader(vertexShaderSrc, gl.VERTEX_SHADER)
		
		if err != nil {
				log.Fatalln("Error:", err)
		}
		fragmentShader, err := compilarShader(fragmentShaderSrc, gl.FRAGMENT_SHADER)

		if err != nil {
				log.Fatalln("Error:", err)
		}
		
		shaderProgram := gl.CreateProgram()
		gl.AttachShader(shaderProgram, vertexShader)
		gl.AttachShader(shaderProgram, fragmentShader)
		gl.LinkProgram(shaderProgram)
		gl.UseProgram(shaderProgram)

		proyeccion := mgl32.Ortho2D(0, float32(config.Ancho), 0, float32(config.Alto))
		posUniforme := gl.GetUniformLocation(shaderProgram, gl.Str("projection\x00"))
		gl.UniformMatrix4fv(posUniforme, 1, false, &proyeccion[0])
		
		//Callback para redimensionar framebuffer, actualizar glViewport, proyeccion de la camara
		window.SetFramebufferSizeCallback(
				func(w *glfw.Window, ancho int, alto int) {
						gl.Viewport(0, 0, int32(ancho), int32(alto))
				
						projection := mgl32.Ortho2D(0, float32(ancho), 0, float32(alto))
						gl.UseProgram(shaderProgram)
						gl.UniformMatrix4fv(posUniforme, 1, false, &projection[0])})

		//ESC para salir, F11 para pantalla completa
		window.SetKeyCallback(
				func(w *glfw.Window, k glfw.Key, scancode int, a glfw.Action, mods glfw.ModifierKey) {
						if a != glfw.Press {
								return
						}
						switch k {
						case glfw.KeyEscape:
								w.SetShouldClose(true)
						case glfw.KeyF11:
								pantallaCompleta(w)
						}
				})

		magnitudes := nuevaMatriz(len(particulas), len(particulas))
		
		min, max := -config.Fuerza, config.Fuerza
		for i := 0; i < len(particulas); i++ {
				for j := i + 1; j < len(particulas); j++ {
						aux := rand.Float64() * (max-min) + min
						for aux == 0 {								
								aux = rand.Float64() * (max-min) + min
						}
						magnitudes[i][j] = aux
				}
		}
		// Loop principal
		for !window.ShouldClose() {
				
				for i := 0; i < len(particulas); i++ {
						for j := i + 1; j < len(particulas); j++ {
								agregarRegla(&particulas[i], particulas[j], magnitudes[i][j])
						}
				}

				// Aplicar amortiguación y actualizar posiciones
				for i := range particulas {
						particulas[i].VX *= 0.99
						particulas[i].VY *= 0.99
						particulas[i].PX += particulas[i].VX
						particulas[i].PY += particulas[i].VY

						//mantener partículas dentro de ventana
						ancho, alto := window.GetSize()
						if particulas[i].PX < 0 { particulas[i].PX = 0; particulas[i].VX *= -1 }
						if particulas[i].PX > float32(ancho) { particulas[i].PX = float32(ancho); particulas[i].VX *= -1 }
						if particulas[i].PY < 0 { particulas[i].PY = 0; particulas[i].VY *= -1 }
						if particulas[i].PY > float32(alto) { particulas[i].PY = float32(alto); particulas[i].VY *= -1 }
				}

				// Reconstruir vértices con las nuevas posiciones
				var vertices []float32
				for _, p := range particulas {
						vertices = append(vertices, p.PX, p.PY, p.ColorR, p.ColorG, p.ColorB)
				}

				// Actualizar el búfer de vértices
				gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
				gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STREAM_DRAW)
				
				gl.ClearColor(0, 0, 0, 1.0)
				gl.Clear(gl.COLOR_BUFFER_BIT)
				gl.BindVertexArray(vao)
				gl.DrawArrays(gl.POINTS, 0, int32(len(particulas)))
				
				// swap buffers y poll events
				window.SwapBuffers()
				glfw.PollEvents()
		}
}


func pantallaCompleta(w *glfw.Window) {
		if !isFullscreen {
				prevX, prevY = w.GetPos()
				pw, ph := w.GetSize()
				prevWidth = pw
				prevHeight = ph

				m := monitor.GetVideoMode()
				w.SetMonitor(monitor, 0, 0, m.Width, m.Height, m.RefreshRate)
				isFullscreen = true
		} else {
				w.SetMonitor(nil, prevX, prevY, prevWidth, prevHeight, 0)
				isFullscreen = false
		}
}


func agregarRegla(p1 *Particula, p2 Particula, m float64){
		var dx, dy float32 = 0,0
		
		dx = p1.PX - p2.PX
		dy = p1.PY - p2.PY
		d := math.Sqrt(float64(dx * dx + dy * dy));
		
		// distancia maxima
		if d > 0 && d < 100000{
				f := float32(m*2/d)
				p1.VX += f * dx
				p1.VY += f * dy
		}
		p1.VX *= 0.5
		p1.VY *= 0.5
		
}

func nuevaMatriz(n, m int) [][]float64 {
    matriz := make([][]float64, n)
    for i := range matriz {
        matriz[i] = make([]float64, m)
    }
    return matriz
}
