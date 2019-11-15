organization := "im.dlg"

name := "dialog-push-service"

version := "0.2.3.0"

scalaVersion := "2.12.10"

crossScalaVersions := List("2.11.11", "2.12.10", "2.13.0")

libraryDependencies ++= Seq(
  "com.thesamet.scalapb" %% "scalapb-runtime" % scalapb.compiler.Version.scalapbVersion % "protobuf",
  "com.thesamet.scalapb" %% "scalapb-runtime-grpc" % scalapb.compiler.Version.scalapbVersion,
  "io.grpc" % "grpc-netty" % scalapb.compiler.Version.grpcJavaVersion
)

PB.targets in Compile := Seq(
  scalapb.gen(singleLineToProtoString = true) → (sourceManaged in Compile).value
)

licenses += ("Apache-2.0", url(
  "https://www.apache.org/licenses/LICENSE-2.0.html"))

publishMavenStyle := true

enablePlugins(Publishing)
