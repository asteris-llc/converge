# This is an example of using the platform module to get version information 

# from the operating system

param "filename" {
  default = "platform.txt"
}

file.content "platformData" {
  destination = "{{ platform.OS }}-{{param `filename`}}"
  content     = "Detected {{ platform.Name }} ({{ platform.OS }}) {{ platform.Version}} {{ platform.LinuxDistribution}}"
}
