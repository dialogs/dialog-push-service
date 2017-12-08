package im.dlg.push.service

import io.grpc._
import io.grpc.stub.StreamObserver
import im.dlg.push.service.push_service._
import scala.concurrent.Future

final case class PushService(host: String, port: Int) {
  private val channel =
    ManagedChannelBuilder.forAddress(host, port).build

  private val asyncStub = PushingGrpc.stub(channel)

  def ping(): Future[PongResponse] =
    asyncStub.ping(PingRequest())

  def stream(failures: StreamObserver[Response]): StreamObserver[Push] =
    asyncStub.pushStream(failures)

  def gracefulShutdown() = channel.shutdown()

  def forcedShutdown() = channel.shutdownNow()

  def isShutdown: Boolean = channel.isShutdown()

  def isTerminated: Boolean = channel.isTerminated()
}
