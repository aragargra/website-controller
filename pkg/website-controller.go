package main

import (
        "fmt" //para formatear cadenas y manejar la salida
        "net/http" //para realizar solicitudes HTTP
        "encoding/json" //para codificar y decodificar JSON
        "io" //para manejo de entrada/salida
        "log" // para el registro de logs
        "mi-app/pkg/v1" // Paquete donde definimos el tipo Website
        "io/ioutil" //para operaciones de entrada/salida, como leer archivos
        "strings" //para manipulación de cadenas
)

func main() {
    log.Println("Iniciando controlador...")
    for {
            //Le indicamos al API Server que queremos recibir actualizaiones de los eventos del recurso Website
            resp, err := http.Get("http://localhost:8001/apis/extensions.example.com/v1/websites?watch=true")
            if err != nil {
                    panic(err)
            }
            defer resp.Body.Close()

            //Leemos la respuesta del API server con os eventos
            decoder := json.NewDecoder(resp.Body)
            for {
                    var event v1.WebsiteWatchEvent
                    if err := decoder.Decode(&event); err == io.EOF {
                            break
                    } else if err != nil {
                            log.Fatal(err)
                    }

                    log.Printf("Recibido el evento: %s: %s: %s\n", event.Type, event.Object.Metadata.Name, event.Object.Spec.GitRepo)

                    //En función del tipo de eventos llamamos a una función o a otra
                    if event.Type == "ADDED" {
                            createWebsite(event.Object)
                    } else if event.Type == "DELETED" {
                            deleteWebsite(event.Object)
                    }
            }
    }

}

//Función para crear el Deployment y el Secret según unas determinadas plantillas (templates)
func createWebsite(website v1.Website) {
    createResource(website, "api/v1", "services", "service-template.json")
    createResource(website, "apis/apps/v1", "deployments", "deployment-template.json")
}

//Función para eliminar el Deployment y el Secret asociados a un Website
func deleteWebsite(website v1.Website) {
    deleteResource(website, "api/v1", "services", getName(website));
    deleteResource(website, "apis/apps/v1", "deployments", getName(website));
}

//Función para crear un recurso
func createResource(webserver v1.Website, apiGroup string, kind string, filename string) {
    log.Printf("Creando %s con el nombre %s en el namespace %s", kind, getName(webserver), webserver.Metadata.Namespace)
    //Leemos la plantilla:
    templateBytes, err := ioutil.ReadFile(filename)
    if err != nil {
            log.Fatal(err)
    }
    //Reemplazamos los marcadores de la plantilla
    //[NAME] se reemplaza con el nombre del recurso (obtenido de getName(webserver))
    template := strings.Replace(string(templateBytes), "[NAME]", getName(webserver), -1)
    //[GIT-REPO] se reemplaza con la URL del repositorio Git, que está en webserver.Spec.GitRepo
    template = strings.Replace(template, "[GIT-REPO]", webserver.Spec.GitRepo, -1)
    //[PUERTO] se reemplaza con el puerto, que se encuentra en webserver.Spec.Puerto
    template = strings.Replace(template, "[PUERTO]", fmt.Sprintf("%d", webserver.Spec.Puerto), -1)
    //log.Printf("Template: %s", template)

    //Se envía una solicitud HTTP POST al servidor API de Kubernetes para crear el recurso
    resp, err := http.Post(fmt.Sprintf("http://localhost:8001/%s/namespaces/%s/%s/", apiGroup, webserver.Metadata.Namespace, kind), "application/json", strings.NewReader(template))
	if err != nil {
            log.Fatal(err)
    }
    log.Println("response Status:", resp.Status)
}

//Función para eliminar un recurso
func deleteResource(webserver v1.Website, apiGroup string, kind string, name string) {
    log.Printf("Eliminando %s de nombre %s en el namespace %s", kind, name, webserver.Metadata.Namespace)
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("http://localhost:8001/%s/namespaces/%s/%s/%s", apiGroup, webserver.Metadata.Namespace, kind, name), nil)
    if err != nil {
            log.Fatal(err)
            return
    }
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
            log.Fatal(err)
            return
    }
    log.Println("Estado de la respuesta:", resp.Status)

}

func getName(website v1.Website) string {
    return website.Metadata.Name + "-website";
}
