import { Controller, Post, Body, HttpCode, HttpStatus } from '@nestjs/common';
import { AppService } from './app.service';

@Controller('auth')
export class AppController {
  constructor(private readonly appService: AppService) {}

  @Post('login')
  @HttpCode(HttpStatus.OK)
  async login(@Body() body: { student_id: string; password?: string }) {
    const password = body.password || 'password123';
    return this.appService.login(body.student_id, password);
  }
}
