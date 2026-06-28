import { Injectable, OnModuleInit, InternalServerErrorException, UnauthorizedException } from '@nestjs/common';
import { JwtService } from '@nestjs/jwt';
import { Pool } from 'pg';
import * as fs from 'fs';
import * as bcrypt from 'bcrypt';

@Injectable()
export class AppService implements OnModuleInit {
  private pool: Pool;

  constructor(private readonly jwtService: JwtService) {
    const connectionString = process.env.DATABASE_URL || 'postgres://postgres:postgres@postgres:5432/auth_db?sslmode=disable';
    this.pool = new Pool({ connectionString });
  }

  async onModuleInit() {
    let connected = false;
    for (let i = 0; i < 5; i++) {
      try {
        await this.pool.query('SELECT 1');
        connected = true;
        break;
      } catch (err) {
        console.log(`Waiting for postgres... attempt ${i + 1}/5: ${err.message}`);
        await new Promise((resolve) => setTimeout(resolve, 2000));
      }
    }

    if (!connected) {
      throw new Error('Failed to connect to database in Auth Service');
    }

    // Create a users table if not exists
    await this.pool.query(`
      CREATE TABLE IF NOT EXISTS users (
        student_id VARCHAR(50) PRIMARY KEY,
        password_hash VARCHAR(100) NOT NULL
      )
    `);

    // Seed dummy users if table is empty
    const res = await this.pool.query('SELECT COUNT(*) FROM users');
    if (parseInt(res.rows[0].count) === 0) {
      console.log('Seeding users...');
      let usersToSeed = ['STUD0014859', 'STUD0015018'];
      
      const csvPath = '/app/ISTN101_unique_users.csv';
      if (fs.existsSync(csvPath)) {
        try {
          const content = fs.readFileSync(csvPath, 'utf8');
          const lines = content.split('\n');
          const csvUsers = [];
          for (let i = 1; i < lines.length; i++) {
            const val = lines[i].trim();
            if (val && val !== 'user_id') {
              csvUsers.push(val);
            }
          }
          if (csvUsers.length > 0) {
            usersToSeed = csvUsers;
            console.log(`Found ${csvUsers.length} users from CSV to seed.`);
          }
        } catch (e) {
          console.error(`Failed to read unique users CSV: ${e.message}`);
        }
      }

      const defaultPassword = 'password123';
      const hash = await bcrypt.hash(defaultPassword, 10);
      
      const client = await this.pool.connect();
      try {
        await client.query('BEGIN');
        for (const id of usersToSeed) {
          await client.query(
            'INSERT INTO users (student_id, password_hash) VALUES ($1, $2) ON CONFLICT DO NOTHING',
            [id, hash]
          );
        }
        await client.query('COMMIT');
        console.log(`Seeded ${usersToSeed.length} users successfully.`);
      } catch (err) {
        await client.query('ROLLBACK');
        console.error(`Failed to seed users: ${err.message}`);
      } finally {
        client.release();
      }
    }
  }

  async login(studentId: string, passwordReq: string) {
    try {
      const res = await this.pool.query('SELECT * FROM users WHERE student_id = $1', [studentId]);
      if (res.rows.length === 0) {
        // If not exists, dynamically register them to make it easy for load testing
        const defaultPassword = 'password123';
        const hash = await bcrypt.hash(defaultPassword, 10);
        await this.pool.query('INSERT INTO users (student_id, password_hash) VALUES ($1, $2)', [studentId, hash]);
        
        // Return JWT token
        const payload = { student_id: studentId };
        return {
          access_token: this.jwtService.sign(payload),
        };
      }

      const user = res.rows[0];
      // Bcrypt hash verification creates high CPU load
      const isMatch = await bcrypt.compare(passwordReq, user.password_hash);
      if (!isMatch) {
        throw new UnauthorizedException('Invalid credentials');
      }

      const payload = { student_id: user.student_id };
      return {
        access_token: this.jwtService.sign(payload),
      };
    } catch (error) {
      if (error instanceof UnauthorizedException) {
        throw error;
      }
      throw new InternalServerErrorException(error.message);
    }
  }
}
